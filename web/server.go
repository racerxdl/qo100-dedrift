package web

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/config"
	"github.com/racerxdl/qo100-dedrift/metrics"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	MessageTypeMainFFT uint8 = iota
	MessageTypeSegFFT        = iota
)

const (
	ClientTimeout = 5 * time.Second
	pongWait      = 60 * time.Second
	writeWait     = 2 * time.Second
)

var log = slog.Scope("Server")

type wsClient struct {
	sync.Mutex
	connection    *websocket.Conn
	closeChan     chan bool
	lastKeepAlive time.Time
}

type Server struct {
	address string

	running      bool
	stopChan     chan bool
	listener     net.Listener
	upgrader     websocket.Upgrader
	clients      []*wsClient
	cLock        sync.Mutex
	maxWsClients int
	settings     []byte
}

func MakeWebServer(address string, maxWsClients int, settings config.WebSettings) *Server {
	s, _ := json.MarshalIndent(settings, "", "   ")
	return &Server{
		address:  address,
		running:  false,
		stopChan: make(chan bool, 1),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:      make([]*wsClient, 0),
		cLock:        sync.Mutex{},
		maxWsClients: maxWsClients,
		settings:     s,
	}
}

func (ws *Server) putClient(c *wsClient) {
	ws.cLock.Lock()

	ws.clients = append(ws.clients, c)

	ws.cLock.Unlock()
}

func (ws *Server) removeClient(c *wsClient) {
	ws.cLock.Lock()

	for i, v := range ws.clients {
		if v == c {
			ws.clients = append(ws.clients[:i], ws.clients[i+1:]...)
			break
		}
	}

	ws.cLock.Unlock()
}

func (ws *Server) maxClients() bool {
	ws.cLock.Lock()
	clientCount := len(ws.clients)
	ws.cLock.Unlock()

	return clientCount >= ws.maxWsClients
}

func (ws *Server) websocket(w http.ResponseWriter, r *http.Request) {
	if ws.maxClients() {
		w.WriteHeader(503)
		_, _ = w.Write([]byte("Max connections reached"))
		return
	}

	c, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error upgrading: %s", err)
		return
	}

	client := &wsClient{
		connection:    c,
		closeChan:     make(chan bool, 1),
		lastKeepAlive: time.Now(),
	}

	ws.putClient(client)
	log.Info("Websocket Client %s connected", c.RemoteAddr())
	running := true
	ticker := time.NewTicker(time.Second)

	metrics.WebConnections.Inc()

	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go func() {
		for running {
			_, data, err := c.ReadMessage()
			if err == nil {
				if len(data) > 0 {
					if string(data) == "KEEP" {
						client.lastKeepAlive = time.Now()
					}
				}
			} else {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Error(err)
				}
				running = false
			}
		}
	}()

	for running {
		select {
		case <-client.closeChan:
			running = false
		case <-ticker.C:
			client.Lock()
			if time.Since(client.lastKeepAlive) > ClientTimeout {
				log.Warn("Client %s timeout.", c.RemoteAddr())
				running = false
			}
			c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("Error sending ping: %s", err)
				running = false
			}
			client.Unlock()
		}
	}
	log.Info("Websocket Client %s disconnected", c.RemoteAddr())
	ticker.Stop()
	ws.removeClient(client)

	// Send Close
	c.SetWriteDeadline(time.Now().Add(time.Second))
	_ = c.WriteMessage(websocket.CloseNormalClosure, nil)

	// Close
	c.Close()
	metrics.WebConnections.Dec()
}

func (ws *Server) BroadcastFFT(fftType uint8, fft []float32) {
	ws.cLock.Lock()

	b := bytes.NewBuffer(nil)
	_ = binary.Write(b, binary.LittleEndian, &fft)

	ob := append([]byte{fftType}, b.Bytes()...) // First Byte is FFT Type

	msg, err := websocket.NewPreparedMessage(websocket.BinaryMessage, ob)

	if err != nil {
		log.Error("Error creating message: %s", err)
		ws.cLock.Unlock()
		return
	}

	for _, v := range ws.clients {
		v.Lock()
		err := v.connection.WritePreparedMessage(msg)
		v.Unlock()
		if err != nil {
			if !strings.Contains(err.Error(), "close sent") {
				log.Error(err)
			}
			v.closeChan <- true
		}
	}

	ws.cLock.Unlock()
}

func (ws *Server) Start() error {
	if ws.running {
		return fmt.Errorf("already running")
	}

	l, err := net.Listen("tcp", ws.address)
	if err != nil {
		return err
	}
	ws.listener = l

	log.Info("Server at %s", ws.address)
	ws.running = true
	go ws.loop()

	return nil
}

func (ws *Server) Stop() {
	if ws.running {
		ws.running = false
		_ = ws.listener.Close()
		for _, v := range ws.clients {
			v.closeChan <- true
		}
		<-ws.stopChan
	}
}

func (ws *Server) loop() {
	srv := &http.Server{}
	router := mux.NewRouter()
	srv.Handler = router

	files := AssetNames()

	for _, f := range files {
		urlPath := path.Join("/", f)
		log.Debug("Registering file %s", urlPath)
		router.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {

			data, err := Asset(urlPath[1:])
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte("Internal Server Error"))
				return
			}

			ext := path.Ext(urlPath)
			mimeType := mime.TypeByExtension(ext)

			if mimeType == "" {
				mimeType = mime.TypeByExtension(".bin")
			}

			w.Header().Add("content-type", mimeType)
			w.WriteHeader(200)
			_, _ = w.Write(data)
		})
	}

	router.Handle("/metrics", metrics.GetHandler())
	router.HandleFunc("/ws", ws.websocket)
	router.HandleFunc("/settings.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/json")
		w.WriteHeader(200)
		_, _ = w.Write(ws.settings)
	})

	indexHandler := func(w http.ResponseWriter, r *http.Request) {
		data, err := Asset("index.html")
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Internal Server Error"))
			return
		}

		w.WriteHeader(200)
		_, _ = w.Write(data)
	}

	router.HandleFunc("/", indexHandler)
	router.NotFoundHandler = http.HandlerFunc(indexHandler)

	err := srv.Serve(ws.listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed") {
		log.Error(err)
	}

	ws.running = false
	ws.stopChan <- true
}
