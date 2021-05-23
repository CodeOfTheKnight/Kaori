package kaoriJwt

import (
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"google.golang.org/api/iterator"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"net/http"
	"strings"
	"time"
)

/*
type JWTContainer struct {
	Token  string
	Fields map[string]interface{}
}

func GenerateTokenPair(db *kaoriDatabase.NoSqlDb, email string, jwtATExpire int) (map[string]JWTContainer, error) {

	//TODO: Generate Access token

	//TODO: Generate Refresh token

	return map[string]JWTContainer{
		"AccessToken":  {token, atClaims},
		"RefreshToken": {refreshToken, rtClaims},
	}, nil
}
*/

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

func CheckOldsToken(db *kaoriDatabase.NoSqlDb, email string) error {

	q := db.Client.C.Collection("User").Doc(email).
		Collection("RefreshToken").
		Where("exp", "<=", time.Now().Unix())

	iter := q.Documents(db.Client.Ctx)
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
		doc.Ref.Delete(db.Client.Ctx)
	}

	return nil
}
