package middleware

import (
	"github.com/weishi258/http-log-collector/log"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status  int
	length  int
	content []byte
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	if w.content == nil {
		w.content = make([]byte, len(b))
		copy(w.content, b)
	} else {
		temp := make([]byte, len(b)+len(w.content))
		copy(temp, w.content)
		copy(temp[len(w.content):], b)
	}
	return n, err
}

func (w *statusWriter) Status() int {
	return w.status
}

func Logger(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := statusWriter{ResponseWriter: w}

		handler(&sw, r)

		end := time.Now()
		latency := end.Sub(start)

		log.GetLogger().Debug(r.RequestURI,
			zap.Int("status", sw.Status()),
			zap.ByteString("response", sw.content),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("form", r.PostForm.Encode()),
			zap.Duration("latency", latency),
		)
	})

}
