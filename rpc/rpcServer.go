package rpc

import (
	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/quan-to/slog"
	"google.golang.org/grpc"
	"net/http"
	"sync"
)

var log = slog.Scope("RPC")

type Server struct {
	sessions    map[string]SessionData
	sessionLock sync.Mutex
	username    string
	password    string
	rc          RemoteControl
	grpc        *grpc.Server
	wrapped     *grpcweb.WrappedGrpcServer
}

func MakeRPCServer(username, password string, rc RemoteControl) *Server {
	s := &Server{
		username: username,
		password: password,
		sessions: make(map[string]SessionData),
		rc:       rc,
		grpc:     grpc.NewServer(),
	}

	RegisterConfigurationServer(s.grpc, s)

	return s
}

func (s *Server) RegisterURLs(r *mux.Router) {
	resources := grpcweb.ListGRPCResources(s.grpc)
	for _, resource := range resources {
		r.HandleFunc(resource, s.ServeHTTP)
	}
}

func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	s.wrapped.ServeHTTP(resp, req)
}

func (s *Server) IsGrpcWebRequest(req *http.Request) bool {
	return s.wrapped.IsGrpcWebRequest(req)
}

func (s *Server) IsGrpcWebsocketRequest(req *http.Request) bool {
	return s.wrapped.IsGrpcWebSocketRequest(req)
}
