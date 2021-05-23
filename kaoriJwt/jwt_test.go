package kaoriJwt

import (
	"testing"
	"time"
)

func TestJWTokenPair_GenerateTokenPair(t *testing.T) {

	jtp := NewJWTokenPair(15, 15448)

	//Set Access Token
	j := jtp.Access.Obj
	j.Iss = "KaoriStream.com"
	j.Iat = time.Now().Unix()
	j.Company = "CodeOfTheKnight"
	j.Email = "prova@gmail.com"
	j.Permission = "ucta"

	//Set RefreshToken
	jtp.Refresh.Obj.Email = "prova@gmail.com"

	err := jtp.GenerateTokenPair("Secret11", "RefreshSecret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(jtp)
}

func TestTokenValid(t *testing.T) {

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InByb3ZhQGdtYWlsLmNvbSIsImV4cCI6MTYyMjY4NzIwNiwicmVmcmVzaElkIjoiMXN3VjlOZ2xvTWxiMEtORE5GVThZOTlzTUswIn0.QsExz1WEBxRaLNV1nYYc468SX3BsVeuD6zhf1P3-rjk"

	err := TokenValid(token, "RefreshSecret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")
}
