// Package logger, Voltium'un access ve error loglarını dosyaya yazar.
// Dosyalar SIGHUP ile yeniden açılabilir (logrotate uyumu, PRD §10.3).
package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// fileWriter, tek bir log dosyasına thread-safe yazar ve Reopen destekler.
type fileWriter struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

func newFileWriter(path string) (*fileWriter, error) {
	fw := &fileWriter{path: path}
	if err := fw.open(); err != nil {
		return nil, err
	}
	return fw, nil
}

func (fw *fileWriter) open() error {
	f, err := os.OpenFile(fw.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	fw.f = f
	return nil
}

func (fw *fileWriter) writeLine(line string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.f == nil {
		return
	}
	fmt.Fprintln(fw.f, line)
}

// Reopen, mevcut dosyayı kapatıp yolu yeniden açar (logrotate sonrası).
func (fw *fileWriter) Reopen() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.f != nil {
		fw.f.Close()
	}
	return fw.open()
}

func (fw *fileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.f == nil {
		return nil
	}
	err := fw.f.Close()
	fw.f = nil
	return err
}

// ErrorLogger, uygulama/hata loglarını nginx error.log formatında yazar.
type ErrorLogger struct {
	fw *fileWriter
}

// NewErrorLogger, verilen yola yazan bir error logger döner.
func NewErrorLogger(path string) (*ErrorLogger, error) {
	fw, err := newFileWriter(path)
	if err != nil {
		return nil, err
	}
	return &ErrorLogger{fw: fw}, nil
}

func (l *ErrorLogger) log(level, msg string) {
	// 2026/04/07 10:00:00 [error] mesaj
	ts := time.Now().Format("2006/01/02 15:04:05")
	l.fw.writeLine(fmt.Sprintf("%s [%s] %s", ts, level, msg))
}

// Error, hata seviyesinde log yazar.
func (l *ErrorLogger) Error(msg string) { l.log("error", msg) }

// Warn, uyarı seviyesinde log yazar.
func (l *ErrorLogger) Warn(msg string) { l.log("warn", msg) }

// Info, bilgi seviyesinde log yazar.
func (l *ErrorLogger) Info(msg string) { l.log("info", msg) }

// Errorf, biçimlendirilmiş hata logu yazar.
func (l *ErrorLogger) Errorf(format string, a ...any) { l.Error(fmt.Sprintf(format, a...)) }

// Reopen, error log dosyasını yeniden açar.
func (l *ErrorLogger) Reopen() error { return l.fw.Reopen() }

// Close, error log dosyasını kapatır.
func (l *ErrorLogger) Close() error { return l.fw.Close() }
