package runner

import "sync"

// TailWriter keeps the last N bytes written. It keeps stderr streaming by acting
// as a Tee target alongside the terminal writer.
type TailWriter struct {
	limit int
	mu    sync.Mutex
	buf   []byte
}

func NewTailWriter(limit int) *TailWriter {
	return &TailWriter{limit: limit}
}

func (t *TailWriter) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(p) >= t.limit {
		t.buf = append([]byte{}, p[len(p)-t.limit:]...)
		return len(p), nil
	}

	t.buf = append(t.buf, p...)
	if len(t.buf) > t.limit {
		t.buf = t.buf[len(t.buf)-t.limit:]
	}
	return len(p), nil
}

func (t *TailWriter) Bytes() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]byte{}, t.buf...)
}
