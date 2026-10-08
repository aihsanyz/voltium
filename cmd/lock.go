package cmd

import (
	"fmt"
	"os"
	"syscall"
)

// acquireLock, data-dir'deki lock dosyasını flock (LOCK_EX|LOCK_NB) ile kilitler.
// Aynı data-dir üzerinde ikinci bir Voltium instance'ını engeller. Kilit, süreç
// sonlanınca otomatik bırakılır; dönen fonksiyon erken bırakma için kullanılabilir.
func acquireLock(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("başka bir Voltium instance çalışıyor gibi görünüyor (lock: %s)", path)
	}
	release := func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
	return release, nil
}
