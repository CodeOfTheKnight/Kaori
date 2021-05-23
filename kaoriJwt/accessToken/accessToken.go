package accessToken

import (
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriJwt"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"github.com/form3tech-oss/jwt-go"
	"strconv"
	"time"
)

type JWTMetadata struct {
	Iss        string
	Iat        int64
	Exp        int64
	Company    string
	Email      string
	Permission string
}

func NewToken(addExpire int) *JWTMetadata {
	return &JWTMetadata{Exp: time.Now().Add(time.Minute * time.Duration(addExpire)).Unix()}
}

func (jm *JWTMetadata) GenerateToken(jwtPass string) (string, error) {

	if err := jm.check(); err != nil {
		return "", err
	}

	/*
	//Get permissions
	document, err := db.Client.C.Collection("User").Doc(jam.Email).Get(db.Client.Ctx)
	if err != nil {
		return "", err
	}
	data := document.Data()
	 */

	//Generate Access token
	atClaims := jwt.MapClaims{}
	atClaims["iss"] = jm.Iss
	atClaims["iat"] = time.Now().Unix()
	atClaims["exp"] = jm.Exp
	atClaims["company"] = jm.Company
	atClaims["email"] = jm.Email
	atClaims["permission"] = jm.Permission
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(jwtPass))
	if err != nil {
		return "", err
	}

	return token, nil
}

func ExtractMetadata(tokenString, secret string) (*JWTMetadata, error) {

	token, err := kaoriJwt.VerifyToken(tokenString, secret)
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

		return &JWTMetadata{
			Iss:        iss,
			Iat:        iat,
			Exp:        exp,
			Company:    company,
			Email:      email,
			Permission: perm,
		}, nil
	}
	return nil, err
}

func (jm *JWTMetadata) check() error {

	if jm.Iss == "" {
		return errors.New("Iss not setted")
	}

	if jm.Iat == 0 {
		return errors.New("Iat not setted")
	}

	if jm.Exp < time.Now().Unix() {
		return errors.New("Exp not valid")
	}

	if jm.Company == "" {
		return errors.New("Company not setted")
	}

	if !kaoriUtils.IsEmailValid(jm.Email) {
		return errors.New("Email not valid")
	}

	if jm.Permission == "" {
		return errors.New("Permission not setted")
	}

	return nil
}