package server

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Server struct {
	http.Server
	Logger *logrus.Logger
}

func (s *Server) ListenAndServe() error {
	s.Logger.Info("http server start listen to ", s.Addr)
	return s.Server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.Info("http server shutdown gracefully")
	return s.Server.Shutdown(ctx)
}
