package main

import (
	"errors"
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"google.golang.org/api/iterator"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type JWTAccessMetadata struct {
	Iss string
	Iat int64
	Exp int64
	Company string
	Email string
	Permission string
}

type JWTRefreshMetadata struct {
	RefreshId string
	Email string
	Exp int64
}

type JWTContainer struct {
	Token string
	Fields map[string]interface{}
}

func GenerateTokenPair(email string) (map[string]JWTContainer, error) {

	expire := time.Now().Add(time.Minute * 3).Unix()

	//Get permissions
	document, err := kaoriUser.Client.c.Collection("User").Doc(email).Get(kaoriUser.Client.ctx)
	if err != nil {
		return nil, err
	}
	data := document.Data()

	//Generate Access token
	atClaims := jwt.MapClaims{}
	atClaims["iss"] = "KaoriStream.com"
	atClaims["iat"] = time.Now().Unix()
	atClaims["exp"] = expire
	atClaims["company"] = "CodeOfTheKnight"
	atClaims["email"] = email
	atClaims["permission"]= data["Permission"].(string)
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	//Generate refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refreshId"] = GenerateID()
	rtClaims["email"] = email
	rtClaims["exp"] = time.Now().Add(168 * time.Hour) //7 days
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	return map[string]JWTContainer{
		"AccessToken": {token, atClaims},
		"RefreshToken": {refreshToken, rtClaims},
	}, nil
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

func VerifyToken(tokenString string, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil

	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func VerifyExpireDate(exp int64) bool {
	if exp >= time.Now().Unix() {
		return true
	}
	return false
}

func VerifyRefreshToken(email, idRefresh string) bool {

	_, err := kaoriUser.Client.c.Collection("User").Doc(email).
										Collection("RefreshToken").Doc(idRefresh).
										Get(kaoriUser.Client.ctx)
	if err != nil {
		return false
	}

	return true
}

func TokenValid(tokenString, secret string) error {
	token, err := VerifyToken(tokenString, secret)
	fmt.Println("TOKEN", err)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

func ExtractAccessTokenMetadata(tokenString, secret string) (*JWTAccessMetadata, error) {
	token, err := VerifyToken(tokenString, secret)
	fmt.Println(token)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		iss, ok := claims["iss"].(string)
		if !ok {
			return nil, err
		}

		iat, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["iat"]), 10, 64)
		if err != nil {
			return nil, err
		}


		exp, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["exp"]), 10, 64)
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


		perm, ok := claims["permission"].(string)
		if !ok {
			return nil, err
		}

		return &JWTAccessMetadata{
			Iss:       	iss,
			Iat:        iat,
			Exp:        exp,
			Company:    company,
			Email:      email,
			Permission: perm,
		}, nil
	}
	return nil, err
}

func ExtractRefreshTokenMetadata(tokenString, secret string) (*JWTRefreshMetadata, error) {
	token, err := VerifyToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		fmt.Println(claims)

		refreshId, ok := claims["refreshId"].(string)
		if !ok {
			return nil, errors.New("RefreshId error!")
		}

		exp := dateToUnix(claims["exp"].(string))

		email, ok := claims["email"].(string)
		if !ok {
			return nil, errors.New("Email error!")
		}

		return &JWTRefreshMetadata{
			RefreshId: refreshId,
			Email:     email,
			Exp:       exp,
		}, nil
	}
	return nil, err
}

func CheckOldsToken(email string) error {
	q := kaoriUser.Client.c.Collection("User").Doc(email).
		Collection("RefreshToken").
		Where("exp", "<=", time.Now().Unix())

	iter := q.Documents(kaoriUser.Client.ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(doc.Data())
		doc.Ref.Delete(kaoriUser.Client.ctx)
	}

	return nil
}
