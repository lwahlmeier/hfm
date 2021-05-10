package main

import (
	"net/http"
	"os"
	"testing"
)

type FakeResponseWriter struct {
	header     http.Header
	body       []byte
	statusCode int
}

func (fhrw *FakeResponseWriter) Header() http.Header {
	return fhrw.header
}

func (fhrw *FakeResponseWriter) Write(ba []byte) (int, error) {
	// fhrw.body = append(fhrw.body, ba...)
	return len(ba), nil
}

func (fhrw *FakeResponseWriter) WriteHeader(statusCode int) {
	fhrw.statusCode = statusCode
}

func Test_dirList(t *testing.T) {
	frw := &FakeResponseWriter{header: make(http.Header)}
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	f, _ := os.Open("/proc")
	dirList(frw, req, f)
	// fmt.Println(string(frw.body))
}

func BenchmarkDirList(b *testing.B) {
	f, _ := os.Open("/proc")
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	frw := &FakeResponseWriter{header: make(http.Header)}
	for i := 0; i < b.N; i++ {

		dirList(frw, req, f)
	}
}
