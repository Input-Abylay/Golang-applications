package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging required: timestamp, http method, endpoint name :contentReference[oaicite:15]{index=15}
func Logging(next http.Handler, message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Path
		if r.URL.RawQuery != "" {
			endpoint += "?" + r.URL.RawQuery
		}
		log.Printf("%s %s %s %s", time.Now().Format(time.RFC3339), r.Method, endpoint, message)
		next.ServeHTTP(w, r)
	})
}
