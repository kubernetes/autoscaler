package testing

import (
	"io"
	"sync"
)

// ByteLoop provides an io.ReadCloser that always fills the read byte slice
// with repeating value, until closed.
type ByteLoop struct {
	value  byte
	closed bool
	mu     sync.RWMutex
}

// Read populates the passed in byte slice with the value the ByteLoop was
// created with. Returns the size of bytes written. If the reader is closed,
// io.EOF will be returned.
func (l *ByteLoop) Read(p []byte) (size int, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.closed {
		return 0, io.EOF
	}

	for i := 0; i < len(p); i++ {
		p[i] = l.value
	}
	return len(p), nil
}

// Close closes the ByteLoop, and prevents any further reading.
func (l *ByteLoop) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true

	return nil
}
