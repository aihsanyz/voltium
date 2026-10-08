package logger

import (
	"fmt"
	"strconv"
	"time"
)

// AccessEntry, tek bir access log satırının alanlarını taşır.
type AccessEntry struct {
	RemoteAddr   string        // $remote_addr (port'suz IP)
	Time         time.Time     // $time_local
	Method       string        // $request'in parçası
	URI          string        // $request'in parçası (RequestURI)
	Proto        string        // $request'in parçası (HTTP/1.1)
	Status       int           // $status
	BodyBytes    int           // $body_bytes_sent
	Referer      string        // $http_referer
	UserAgent    string        // $http_user_agent
	RequestTime  time.Duration // $request_time
	UpstreamAddr string        // $upstream_addr (ör. 127.0.0.1:3000, "file", "blocked:waf")
}

// AccessLogger, istekleri nginx combined formatında yazar (PRD §10.1).
type AccessLogger struct {
	fw *fileWriter
}

// NewAccessLogger, verilen yola yazan bir access logger döner.
func NewAccessLogger(path string) (*AccessLogger, error) {
	fw, err := newFileWriter(path)
	if err != nil {
		return nil, err
	}
	return &AccessLogger{fw: fw}, nil
}

// Log, bir access satırı yazar.
// Format: $remote_addr - - [$time_local] "$request" $status $body_bytes_sent
//
//	"$http_referer" "$http_user_agent" $request_time $upstream_addr
func (l *AccessLogger) Log(e AccessEntry) {
	l.fw.writeLine(e.format())
}

func (e AccessEntry) format() string {
	request := fmt.Sprintf("%s %s %s", e.Method, e.URI, e.Proto)
	referer := dashIfEmpty(e.Referer)
	ua := dashIfEmpty(e.UserAgent)
	upstream := dashIfEmpty(e.UpstreamAddr)
	timeLocal := e.Time.Format("02/Jan/2006:15:04:05 -0700")
	reqTime := strconv.FormatFloat(e.RequestTime.Seconds(), 'f', 3, 64)

	return fmt.Sprintf(`%s - - [%s] "%s" %d %d "%s" "%s" %s %s`,
		dashIfEmpty(e.RemoteAddr),
		timeLocal,
		request,
		e.Status,
		e.BodyBytes,
		referer,
		ua,
		reqTime,
		upstream,
	)
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// Reopen, access log dosyasını yeniden açar.
func (l *AccessLogger) Reopen() error { return l.fw.Reopen() }

// Close, access log dosyasını kapatır.
func (l *AccessLogger) Close() error { return l.fw.Close() }
