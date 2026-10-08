package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

const shutdownTimeout = 30 * time.Second

// Server, aynı handler'ı birden fazla portta dinleyen HTTP sunucularını yönetir.
type Server struct {
	handler http.Handler
	mu      sync.Mutex
	servers []*http.Server
}

// NewServer, verilen handler için bir Server döner.
func NewServer(handler http.Handler) *Server {
	return &Server{handler: handler}
}

// Start, her port için bir listener açar. Bir port bind edilemezse (ör. kullanımda)
// senkron olarak hata döner; başarıyla açılanlar arka planda servise başlar.
func (s *Server) Start(ports []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range ports {
		addr := fmt.Sprintf(":%d", p)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("port %d dinlenemedi: %w", p, err)
		}
		srv := &http.Server{Handler: s.handler}
		s.servers = append(s.servers, srv)
		go srv.Serve(ln)
	}
	return nil
}

// Shutdown, tüm sunucuları graceful biçimde kapatır (30s deadline).
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.mu.Lock()
	servers := append([]*http.Server(nil), s.servers...)
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, srv := range servers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()
			_ = srv.Shutdown(ctx)
		}(srv)
	}
	wg.Wait()
}
