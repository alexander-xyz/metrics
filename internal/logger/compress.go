package logger

import (
	"compress/gzip"
	"net/http"
	"strings"
)

var compressibleTypes = []string{"application/json", "text/html"}

func isCompressible(contentType string) bool {
	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}

	return false
}

type compressWriter struct {
	http.ResponseWriter
	zw          *gzip.Writer
	compress    bool
	wroteHeader bool
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.wroteHeader = true

		if isCompressible(c.Header().Get("Content-Type")) {
			c.compress = true
			c.Header().Set("Content-Encoding", "gzip")
			c.Header().Del("Content-Length")
		}
	}

	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	if c.compress {
		return c.zw.Write(b)
	}

	return c.ResponseWriter.Write(b)
}

func (c *compressWriter) Close() error {
	if c.compress {
		return c.zw.Close()
	}

	return nil
}

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			zr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "cannot decompress request body", http.StatusBadRequest)
				return
			}
			defer zr.Close()

			r.Body = zr
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		zw := gzip.NewWriter(w)
		cw := &compressWriter{ResponseWriter: w, zw: zw}

		defer func() {
			if err := cw.Close(); err != nil {
				Log.Error(err.Error())
			}
		}()

		h.ServeHTTP(cw, r)
	})
}
