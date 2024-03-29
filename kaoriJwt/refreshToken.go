package kaoriJwt

import (
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"github.com/form3tech-oss/jwt-go"
	"strconv"
	"time"
)

type JWTRefreshMetadata struct {
	RefreshId string
	Email     string
	Exp       int64
}

func NewRefreshToken(addExpire int) *JWTRefreshMetadata {
	return &JWTRefreshMetadata{Exp: time.Now().Add(time.Minute * time.Duration(addExpire)).Unix()}
}

func (jm *JWTRefreshMetadata)  GenerateToken(refreshPass string) (string, error) {

	if err := jm.check(); err != nil {
		return "", err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refreshId"] = jm.RefreshId
	rtClaims["email"] = jm.Email
	rtClaims["exp"] = jm.Exp
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(refreshPass))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func VerifyRefreshToken(db *kaoriDatabase.NoSqlDb, email, idRefresh string) bool {

	_, err := db.Client.C.Collection("User").Doc(email).
		Collection("RefreshToken").Doc(idRefresh).
		Get(db.Client.Ctx)
	if err != nil {
		return false
	}

	return true
}

func ExtractRefreshMetadata(tokenString, secret string) (*JWTRefreshMetadata, error) {

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

		exp, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["exp"]), 10, 64)
		if err != nil {
			return nil, err
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, errors.New("Email error!")
		}

		return &JWTRefreshMetadata {
			RefreshId: refreshId,
			Email:     email,
			Exp:       exp,
		}, nil
	}
	return nil, err
}

func (jm *JWTRefreshMetadata) check() error {

	if !kaoriUtils.IsEmailValid(jm.Email) {
		return errors.New("Email not valid")
	}

	if jm.Exp < time.Now().Unix() {
		return errors.New("Expiration not valid")
	}

	return nil
}