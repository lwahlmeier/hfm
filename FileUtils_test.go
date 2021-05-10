package main

import (
	"net/http"
	"testing"
)

type FakeReader struct {
	ba []byte
}

func (fr *FakeReader) Read(ba []byte) (int, error) {
	copy(fr.ba[:len(ba)], ba)
	return len(ba), nil
}

func Test_copyBufferN(t *testing.T) {
	frw := &FakeResponseWriter{header: make(http.Header)}
	fr := &FakeReader{ba: make([]byte, 1024*1024)}
	copyBufferN(frw, fr, 1024*1024)
}

func BenchmarkCopyBufferN(b *testing.B) {
	frw := &FakeResponseWriter{header: make(http.Header)}
	fr := &FakeReader{ba: make([]byte, 1024*1024)}
	for i := 0; i < b.N; i++ {
		copyBufferN(frw, fr, 1024*1024)
	}
}
