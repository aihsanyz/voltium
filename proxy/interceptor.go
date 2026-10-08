package proxy

import "net/http"

// interceptor, ResponseWriter'ı sarmalayıp status ve gönderilen byte sayısını yakalar.
// WriteHeader anı, Faz-2'de onResponse zincirinin (header-only mutasyon) çalışacağı
// dikiş noktasıdır — böylece Faz-2'de handler yeniden yazılmaz.
type interceptor struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (i *interceptor) WriteHeader(code int) {
	if i.wroteHeader {
		return
	}
	i.status = code
	i.wroteHeader = true
	// Faz-2 dikiş noktası: onResponse zinciri header/status'u burada mutasyona uğratacak.
	i.ResponseWriter.WriteHeader(code)
}

func (i *interceptor) Write(b []byte) (int, error) {
	if !i.wroteHeader {
		i.WriteHeader(http.StatusOK)
	}
	n, err := i.ResponseWriter.Write(b)
	i.bytes += n
	return n, err
}

// Flush, streaming yanıtlar (ReverseProxy) için alttaki Flusher'a delege eder.
func (i *interceptor) Flush() {
	if f, ok := i.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
