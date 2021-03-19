package main

import (
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type JWTMetadata struct {
	Iss string
	Iat uint64
	Exp uint64
	Company string
	Email string
	Password string
	Permission string
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")

	fmt.Println(bearToken)

	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil

	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	fmt.Println("TOKEN", err)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

func ExtractTokenMetadata(r *http.Request) (*JWTMetadata, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		iss, ok := claims["iss"].(string)
		if !ok {
			return nil, err
		}

		iat, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["iat"]), 10, 64)
		if err != nil {
			return nil, err
		}

		exp, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["exp"]), 10, 64)
		if err != nil {
			return nil, err
		}

		company, ok := claims["company"].(string)
		if !ok {
			return nil, err
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, err
		}

		password, ok := claims["password"].(string)
		if !ok {
			return nil, err
		}

		perm, ok := claims["permission"].(string)
		if !ok {
			return nil, err
		}

		return &JWTMetadata{
			Iss:       	iss,
			Iat:        iat,
			Exp:        exp,
			Company:    company,
			Email:      email,
			Password:   password,
			Permission: perm,
		}, nil
	}
	return nil, err
}
