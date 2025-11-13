package main

import (
	"context"
	"net/http"
)

type Server struct {
	server *http.Server
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx *context.Context) error {
	return s.server.Shutdown(*ctx)
}
