package rpc

import (
	"lfts/internal/utils"
	"net/http"
)

// Server represents the RPC server
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new RPC server instance
func NewServer(port string) *Server {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/status", HandleStatus)
	mux.HandleFunc("/block/latest", HandleLatestBlock)
	mux.HandleFunc("/ftso/price", HandleFTSOPrice)
	mux.HandleFunc("/ftso/inject", HandleInjectFTSO)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return &Server{
		httpServer: server,
	}
}

// Start starts the RPC server
func (s *Server) Start() error {
	utils.Info("RPC server starting on port %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// GetHTTPServer returns the underlying HTTP server (for graceful shutdown if needed)
func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}

