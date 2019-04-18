package web

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/metrics"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	MessageTypeMainFFT uint8 = iota
	MessageTypeSegFFT        = iota
)

var log = slog.Scope("Server")

type wsClient struct {
	connection *websocket.Conn
	closeChan  chan bool
}

type Server struct {
	address string

	running  bool
	stopChan chan bool
	listener net.Listener
	upgrader websocket.Upgrader
	clients  []*wsClient
	cLock    sync.Mutex
}

func MakeWebServer(address string) *Server {
	return &Server{
		address:  address,
		running:  false,
		stopChan: make(chan bool, 1),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make([]*wsClient, 0),
		cLock:   sync.Mutex{},
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

func (ws *Server) websocket(w http.ResponseWriter, r *http.Request) {
	c, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error upgrading: %s", err)
		return
	}

	client := &wsClient{
		connection: c,
		closeChan:  make(chan bool, 1),
	}

	ws.putClient(client)
	log.Info("Websocket Client %s connected", c.RemoteAddr())
	running := true
	ticker := time.NewTicker(time.Second)

	for running {
		select {
		case <-client.closeChan:
			running = false
		case <-ticker.C:
			c.SetWriteDeadline(time.Now().Add(time.Second))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("Error sending ping: %s", err)
				running = false
			}
		default:

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
		err := v.connection.WritePreparedMessage(msg)
		if err != nil {
			log.Error(err)
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

	router.Handle("/metrics", metrics.GetHandler())
	router.HandleFunc("/ws", ws.websocket)

	err := srv.Serve(ws.listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed") {
		log.Error(err)
	}

	ws.running = false
	ws.stopChan <- true
}
