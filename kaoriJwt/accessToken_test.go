package kaoriJwt

import (
	"testing"
	"time"
)

func TestJWTAccessMetadata_GenerateToken(t *testing.T) {

	j := NewAccessToken(15)
	j.Iss = "KaoriStream.com"
	j.Iat = time.Now().Unix()
	j.Company = "CodeOfTheKnight"
	j.Email = "prova@gmail.com"
	j.Permission = "ucta"

	token, err := j.GenerateToken("Secret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)

}

func TestExtractAccessMetadata(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb21wYW55IjoiQ29kZU9mVGhlS25pZ2h0IiwiZW1haWwiOiJwcm92YUBnbWFpbC5jb20iLCJleHAiOjE2MjE3ODQ4MDIsImlhdCI6MTYyMTc4MzkwMiwiaXNzIjoiS2FvcmlTdHJlYW0uY29tIiwicGVybWlzc2lvbiI6InVjdGEifQ.kOslUdJtfw5e-l3olvE_WlZENTjKZlL07pJ_cGydZco"

	j, err := ExtractAccessMetadata(token, "Secret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(j)
}
