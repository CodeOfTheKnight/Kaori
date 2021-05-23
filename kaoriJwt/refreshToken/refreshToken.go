package refreshToken

import (
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriJwt"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/form3tech-oss/jwt-go"
	"time"
)

type JWTMetadata struct {
	RefreshId string
	Email     string
	Exp       int64
}

func NewToken(addExpire int) *JWTMetadata {
	return &JWTMetadata{Exp: time.Now().Add(time.Minute * time.Duration(addExpire)).Unix()}
}

func (jm *JWTMetadata)  GenerateToken(refreshPass string) (string, error) {

	if err := jm.check(); err != nil {
		return "", err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refreshId"] = kaoriUtils.GenerateID()
	rtClaims["email"] = jm.Email
	rtClaims["exp"] = jm.Exp
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(refreshPass))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func VerifyToken(db *kaoriDatabase.NoSqlDb, email, idRefresh string) bool {

	_, err := db.Client.C.Collection("User").Doc(email).
		Collection("RefreshToken").Doc(idRefresh).
		Get(db.Client.Ctx)
	if err != nil {
		return false
	}

	return true
}

func ExtractMetadata(tokenString, secret string) (*JWTMetadata, error) {

	token, err := kaoriJwt.VerifyToken(tokenString, secret)
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

		exp := kaoriUtils.DateToUnix(claims["exp"].(string))

		email, ok := claims["email"].(string)
		if !ok {
			return nil, errors.New("Email error!")
		}

		return &JWTMetadata{
			RefreshId: refreshId,
			Email:     email,
			Exp:       exp,
		}, nil
	}
	return nil, err
}

func (jm *JWTMetadata) check() error {

	if !kaoriUtils.IsEmailValid(jm.Email) {
		return errors.New("Email not valid")
	}

	if jm.Exp < time.Now().Unix() {
		return errors.New("Expiration not valid")
	}

	return nil
}