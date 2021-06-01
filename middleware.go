package main

import (
	"context"
	"github.com/CodeOfTheKnight/Kaori/kaoriJwt"
	"github.com/CodeOfTheKnight/Kaori/kaoriLog"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	logger "github.com/sirupsen/logrus"
	"net/http"
)

type AuthMiddlewarePerm struct {
	PermRequired []kaoriJwt.Permission
}

type ContextValues struct {
	m map[string]string
}


func (v ContextValues) Get(key string) string {
	return v.m[key]
}

func NewAuthMiddlewarePerm(perms ...kaoriJwt.Permission) *AuthMiddlewarePerm {
	var amp AuthMiddlewarePerm
	for _, p := range perms {
		amp.PermRequired = append(amp.PermRequired, p)
	}
	return &amp
}

func (amp *AuthMiddlewarePerm) authmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip := kaoriUtils.GetIP(r)

		//Extract token JWT
		tokenString := kaoriJwt.ExtractToken(r)
		metadata, err := kaoriJwt.ExtractAccessMetadata(tokenString, cfg.Password.AccessToken)
		if err != nil {
			kaoriLog.PrintLog("General", ip, "authMiddleware", "Error to get access token metadata: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, "Token not valid")
			return
		}

		permissions, err := metadata.GetPermission()
		if err != nil {
			kaoriLog.PrintLog(metadata.Email, ip, "authMiddleware", "Error to get permissions: "+err.Error(), 1)
			kaoriUtils.PrintInternalErr(w)
			return
		}

		if !kaoriJwt.IsAuthorized(permissions, amp.PermRequired...) {
			kaoriLog.PrintLog(metadata.Email, ip, "authMiddleware", "Warning permission denied!", 2)
			http.Error(w, `{"code": 403, "msg": "You need permissions to access the service!"}`, http.StatusForbidden)
			return
		}

		v := ContextValues{map[string]string{
			"email": metadata.Email,
			"ip": ip,
		}}
		ctx := context.WithValue(r.Context(), "values", v)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func refreshMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip := kaoriUtils.GetIP(r)

		//Get token from cookies
		cookieData, err := kaoriUtils.GetCookies(r, cfg.Jwt.Iss, cfg.Password.Cookies)
		if err != nil {
			kaoriLog.PrintLog("General", ip, "ApiRefresh", "Error to get cookies: "+err.Error(), 1)
			redirect, err := kaoriUtils.ParseTemplate(cfg.Template.Html["redirect"], "https://" + cfg.Server.Host + cfg.Server.Port + endpointLogin.String())
			if err != nil {
				logger.WithFields(logger.Fields{
					"function": "refreshMiddleware",
					"ip": ip,
				}).Error("Unable to parse email template")
				kaoriUtils.PrintInternalErr(w)
				return
			}
			w.Write([]byte(redirect))
			return
		}

		token := cookieData["RefreshToken"]

		//Extract data from token
		data, err := kaoriJwt.ExtractRefreshMetadata(token, cfg.Password.RefreshToken)
		if err != nil {
			kaoriLog.PrintLog("General", ip, "ApiRefresh", "Extract refresh token error: "+err.Error(), 1)
			redirect, err := kaoriUtils.ParseTemplate(cfg.Template.Html["redirect"], cfg.Server.Host + cfg.Server.Port + endpointLogin.String())
			if err != nil {
				logger.WithFields(logger.Fields{
					"function": "refreshMiddleware",
					"ip": ip,
				}).Error("Unable to parse email template")
				kaoriUtils.PrintInternalErr(w)
				return
			}
			w.Write([]byte(redirect))
			return
		}

		//Check validity
		if !kaoriJwt.VerifyRefreshToken(kaoriUserDB, data.Email, data.RefreshId) {
			kaoriLog.PrintLog(data.Email, ip, "ApiRefresh", "Token not valid", 2)
			redirect, err := kaoriUtils.ParseTemplate(cfg.Template.Html["redirect"], cfg.Server.Host + cfg.Server.Port + endpointLogin.String())
			if err != nil {
				logger.WithFields(logger.Fields{
					"function": "refreshMiddleware",
					"ip": ip,
				}).Error("Unable to parse email template")
				kaoriUtils.PrintInternalErr(w)
				return
			}
			w.Write([]byte(redirect))
			return
		}

		if !kaoriJwt.VerifyExpireDate(data.Exp) {
			kaoriLog.PrintLog(data.Email, ip, "ApiRefresh", "Token expired", 2)
			redirect, err := kaoriUtils.ParseTemplate(cfg.Template.Html["redirect"], cfg.Server.Host + cfg.Server.Port + endpointLogin.String())
			if err != nil {
				logger.WithFields(logger.Fields{
					"function": "refreshMiddleware",
					"ip": ip,
				}).Error("Unable to parse email template")
				kaoriUtils.PrintInternalErr(w)
				return
			}
			w.Write([]byte(redirect))
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
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}