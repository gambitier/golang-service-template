package server

import (
	"context"
	"log/slog"

	appitem "github.com/gambitier/golang-service-template/internal/application/item"
	"github.com/gambitier/golang-service-template/internal/config"
	"github.com/gambitier/golang-service-template/internal/presentation/http/handlers"
	httpserver "github.com/gambitier/golang-service-template/internal/server/http"
)

// Server runs the HTTP API.
type Server struct {
	httpServer *httpserver.Server
}

// New wires handlers and the HTTP server.
func New(cfg *config.Config, logger *slog.Logger, itemSvc *appitem.Service) (*Server, error) {
	h := handlers.NewHandlers(itemSvc)
	httpSrv, err := httpserver.New(cfg, logger, h)
	if err != nil {
		return nil, err
	}
	return &Server{httpServer: httpSrv}, nil
}

// Start runs HTTP until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	return s.httpServer.Start(ctx)
}
