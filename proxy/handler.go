package proxy

import (
	"net"
	"net/http"
	"time"

	"voltium/logger"
)

// Handler, istek pipeline'ını yürütür: Host oku → route → redirect/upstream → access log.
// Faz-1'de plugin zinciri yoktur; onRequest/onResponse dikiş noktaları yorumlarla işaretlidir.
type Handler struct {
	router *Router
	access *logger.AccessLogger // nil olabilir (testlerde)
	errlog *logger.ErrorLogger  // nil olabilir
}

// NewHandler, verilen router ve logger'larla bir handler döner.
func NewHandler(router *Router, access *logger.AccessLogger, errlog *logger.ErrorLogger) *Handler {
	return &Handler{router: router, access: access, errlog: errlog}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	iw := &interceptor{ResponseWriter: w}

	entry, ok := h.router.Lookup(r.Host)
	upstreamAddr := "-"

	switch {
	case !ok:
		iw.WriteHeader(http.StatusNotFound)
	case entry.IsRedirect():
		iw.Header().Set("Location", entry.Redirect)
		iw.WriteHeader(entry.RedirectCode)
	default:
		// Faz-2 dikiş: onRequest zinciri burada çalışacak; "block" dönerse short-circuit.
		upstreamAddr = entry.Upstream.Addr()
		entry.Upstream.Forward(iw, r)
		// Faz-2 dikiş: onResponse zinciri interceptor.WriteHeader içinde çalışır.
	}

	if iw.status == 0 {
		iw.status = http.StatusOK
	}

	if h.access != nil {
		h.access.Log(logger.AccessEntry{
			RemoteAddr:   clientHost(r.RemoteAddr),
			Time:         start,
			Method:       r.Method,
			URI:          r.RequestURI,
			Proto:        r.Proto,
			Status:       iw.status,
			BodyBytes:    iw.bytes,
			Referer:      r.Referer(),
			UserAgent:    r.UserAgent(),
			RequestTime:  time.Since(start),
			UpstreamAddr: upstreamAddr,
		})
	}
}

func clientHost(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
