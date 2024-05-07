package router

import (
	"net/http"
	"ultraphx-core/internal/services/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtStr := r.Header.Get("Authorization")
		if jwtStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ok, err := auth.CheckJwtToken(jwtStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
