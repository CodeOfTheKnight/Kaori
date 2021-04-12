package main

import (
	"net/http"
)

type AuthMiddlewarePerm struct {
	PermRequired []Permission
}

func NewAuthMiddlewarePerm(perms ...Permission) *AuthMiddlewarePerm {
	var amp AuthMiddlewarePerm
	for _, p := range perms {
		amp.PermRequired = append(amp.PermRequired, p)
	}
	return &amp
}

func (amp *AuthMiddlewarePerm) authmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip := GetIP(r)

		//Extract token JWT
		tokenString := ExtractToken(r)
		metadata, err := ExtractAccessTokenMetadata(tokenString, cfg.Password.AccessToken)
		if err != nil {
			printLog("General", ip, "authMiddleware", "Error to get access token metadata: "+err.Error(), 1)
			printErr(w, "Token not valid")
			return
		}

		permissions, err := metadata.GetPermission()
		if err != nil {
			printLog(metadata.Email, ip, "authMiddleware", "Error to get permissions: "+err.Error(), 1)
			printInternalErr(w)
			return
		}

		if !IsAuthorized(permissions, amp.PermRequired...) {
			printLog(metadata.Email, ip, "authMiddleware", "Warning permission denied: "+err.Error(), 2)
			http.Error(w, `{"code": 403, "msg": "You need permissions to access the service!"}`, http.StatusForbidden)
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)

	})
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}
