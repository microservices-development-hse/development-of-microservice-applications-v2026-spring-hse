package handler

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logrus.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"remote":   r.RemoteAddr,
			"duration": time.Since(start).String(),
		}).Info("HTTP Request")
	})
}
