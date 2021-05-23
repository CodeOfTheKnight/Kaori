package refreshToken

import "testing"

func TestJWTRefreshMetadata_GenerateToken(t *testing.T) {

	jr := NewToken(15062)
	jr.Email = "prova@gmail.com"

	token, err := jr.GenerateToken("RefreshSecret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)
}

func TestExtractRefreshTokenMetadata(t *testing.T) {



}
