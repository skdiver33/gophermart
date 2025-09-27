package middleware

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
}

func (w gzipWriter) Write(b []byte) (int, error) {
	//use compression if size>100 bytes
	if len(b) > 100 {
		gz, err := gzip.NewWriterLevel(w.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			log.Print("error create gzip compressor for response")
			return w.ResponseWriter.Write(b)
		}
		w.Header().Set("Content-Encoding", "gzip")
		count, err := gz.Write(b)
		gz.Close()
		return count, err
	}
	return w.ResponseWriter.Write(b)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(gzipWriter{ResponseWriter: w}, r)
	})
}
