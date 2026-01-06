package security

import (
	"net/http"
	"os"
)

func InternalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		internalToken := os.Getenv("INTERNAL_TOKEN")
		requestToken := r.Header.Get("X-Internal-Token")

		if requestToken == "" || requestToken != internalToken {
			http.Error(w, "INTRUSO - acesso negado", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
