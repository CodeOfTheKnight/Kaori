package main

import (
	"net/http"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(1, 3)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() == false {
			http.Error(w, `{"code": 429, "msg": "Too many request!"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
