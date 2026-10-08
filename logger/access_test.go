package logger

import (
	"testing"
	"time"
)

func TestAccessEntryFormat(t *testing.T) {
	ts := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	e := AccessEntry{
		RemoteAddr:   "1.2.3.4",
		Time:         ts,
		Method:       "GET",
		URI:          "/api/users",
		Proto:        "HTTP/1.1",
		Status:       200,
		BodyBytes:    1234,
		UserAgent:    "Mozilla/5.0",
		RequestTime:  18 * time.Millisecond,
		UpstreamAddr: "127.0.0.1:3000",
	}
	got := e.format()
	want := `1.2.3.4 - - [07/Apr/2026:10:00:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0" 0.018 127.0.0.1:3000`
	if got != want {
		t.Errorf("format()\n got: %s\nwant: %s", got, want)
	}
}

func TestAccessEntryBlocked(t *testing.T) {
	ts := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	e := AccessEntry{
		RemoteAddr:   "1.2.3.4",
		Time:         ts,
		Method:       "GET",
		URI:          "/api/users",
		Proto:        "HTTP/1.1",
		Status:       403,
		BodyBytes:    0,
		UserAgent:    "sqlmap/1.0",
		RequestTime:  time.Millisecond,
		UpstreamAddr: "blocked:waf",
	}
	want := `1.2.3.4 - - [07/Apr/2026:10:00:00 +0000] "GET /api/users HTTP/1.1" 403 0 "-" "sqlmap/1.0" 0.001 blocked:waf`
	if got := e.format(); got != want {
		t.Errorf("format()\n got: %s\nwant: %s", got, want)
	}
}
