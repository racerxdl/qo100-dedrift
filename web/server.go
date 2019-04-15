package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/metrics"
	"net"
	"net/http"
	"strings"
)

var log = slog.Scope("Server")

type Server struct {
	address string

	running  bool
	stopChan chan bool
	listener net.Listener
}

func MakeWebServer(address string) *Server {
	return &Server{
		address:  address,
		running:  false,
		stopChan: make(chan bool, 1),
	}
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
		<-ws.stopChan
	}
}

func (ws *Server) loop() {
	srv := &http.Server{}
	router := mux.NewRouter()
	srv.Handler = router

	router.Handle("/metrics", metrics.GetHandler())

	err := srv.Serve(ws.listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed") {
		log.Error(err)
	}

	ws.running = false
	ws.stopChan <- true
}
