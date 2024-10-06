package main

import (
	"net/http"
)

func (cfg *apiConfig) metricsIncMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// type contextKey string

// const userIdKey contextKey = "userId"

// func (cfg *apiConfig) authMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		token, err := auth.GetBearerToken(r.Header)
// 		if err != nil {
// 			respError(w, 401, "Couldn't find JWT", err)
// 			return
// 		}

// 		userId, err := auth.ValidateJWT(token, cfg.secretKeyJWT)
// 		if err != nil {
// 			respError(w, 401, "Couldn't validate JWT", err)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), userIdKey, userId)
// 		r = r.WithContext(ctx)

// 		next.ServeHTTP(w, r)
// 	})
// }
