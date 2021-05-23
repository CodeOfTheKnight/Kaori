package kaoriJwt

import "testing"

func TestJWTRefreshMetadata_GenerateToken(t *testing.T) {

	jr := NewAccessToken(15062)
	jr.Email = "prova@gmail.com"

	token, err := jr.GenerateToken("RefreshSecret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)
}

func TestExtractRefreshMetadata(t *testing.T) {

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InByb3ZhQGdtYWlsLmNvbSIsImV4cCI6MTYyMjY4NzIwNiwicmVmcmVzaElkIjoiMXN3VjlOZ2xvTWxiMEtORE5GVThZOTlzTUswIn0.QsExz1WEBxRaLNV1nYYc468SX3BsVeuD6zhf1P3-rjk"

	j, err := ExtractRefreshMetadata(token, "RefreshSecret11")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(j)
}
