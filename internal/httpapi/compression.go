package httpapi

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) { return w.writer.Write(b) }
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		g := gzip.NewWriter(w)
		defer g.Close()
		next.ServeHTTP(gzipWriter{ResponseWriter: w, writer: g}, r)
	})
}
