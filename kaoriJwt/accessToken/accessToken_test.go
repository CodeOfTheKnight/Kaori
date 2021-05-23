package accessToken

import (
	"testing"
	"time"
)

func TestJWTAccessMetadata_GenerateToken(t *testing.T) {

	j := NewToken(15)
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
