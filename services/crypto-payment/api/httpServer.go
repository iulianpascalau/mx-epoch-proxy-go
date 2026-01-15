package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("api")

// HTTPServer defines the specific implementation for the HTTP server
type httpServer struct {
	server *http.Server
}

// NewHTTPServer creates a new instance of httpServer
func NewHTTPServer(handler *Handler, port int) *httpServer {
	router := gin.Default()

	router.GET("/ping", handler.Ping)
	router.POST("/create-address", handler.CreateAddress)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &httpServer{server: server}
}

// Start launches the HTTP server in a separate goroutine
func (s *httpServer) Start() error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}

	// Update the server address to reflect the actual port (useful if port was 0)
	s.server.Addr = listener.Addr().String()

	go func() {
		log.Info("starting HTTP server", "address", s.server.Addr)
		errServe := s.server.Serve(listener)
		if errServe != nil && errServe != http.ErrServerClosed {
			log.Error("HTTP server stopped unexpectedly", "error", errServe)
		}
	}()

	return nil
}

// Close gracefully shuts down the HTTP server
func (s *httpServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
