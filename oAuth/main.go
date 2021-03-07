package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"kirito/anilistgo"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

type DataAuth struct {
	Token_type   string `json:"token_type"`
	Expires_in   int    `json:"expires_in"`
	Access_token string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	http.HandleFunc("/", HomePageHandler)
	http.HandleFunc("/auth", AutenticationCode)
	fmt.Println(">>>>>>> OClient started at:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
	return
}

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AutenticationCode(w http.ResponseWriter, r *http.Request) {

	codeMatrix := strings.Split(r.URL.RequestURI(), "=")
	code := codeMatrix[1]
	log.Println(code)

	client := &http.Client{}

	data := url.Values{}
	data.Set("client_id", "4204")
	data.Set("client_secret", "QI6p3aHYtnXvf063VucVHVXWUta268D1Nr2Mk8LO")
	data.Set("client_secret", "QI6p3aHYtnXvf063VucVHVXWUta268D1Nr2Mk8LO")
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:8000/auth")

	r, _ = http.NewRequest(http.MethodPost, "https://anilist.co/api/v2/oauth/token", strings.NewReader(data.Encode())) // URL-encoded payload
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)

	ritorno, _ := ioutil.ReadAll(resp.Body)

  writeFileJson(ritorno)

  AddElementAnilist(1, "CURRENT")

}

func AddElementAnilist(idAnilist int, state string) {

  var da DataAuth

  //ReadFile Json for token
  readFileJson(&da)

	//Creazione query
	query := `mutation($mediaId: Int, $status: MediaListStatus) {
                  SaveMediaListEntry (mediaId: $mediaId, status: $status) {
                      id
                      status
                  }
              }`
	variables := struct {
		MediaId int    `json:"mediaId"`
		Status  string `json:"status"`
	}{
		idAnilist,
		state,
	}

	qrs := anilistgo.Query{query, variables}

	jsonQ, err := json.Marshal(qrs)
	if err != nil {
		fmt.Println(err)
	}

	//Creazione richiesta
	req, err := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewReader([]byte(jsonQ)))
	if err != nil {
		fmt.Println(err)
	}

	//Settaggio header
	req.Header.Set("Authorization", "Bearer " + da.Access_token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	//creazione client
	client2 := &http.Client{}

	//Making request
	resp, err := client2.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	response, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(response))

}

func writeFileJson(resp []byte) {

  var da DataAuth

  err := json.Unmarshal(resp, &da)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println(string(da.Access_token))

  b, err := json.MarshalIndent(da, "", "\t")
       if err != nil {
           fmt.Println("error:", err)
  }

  ioutil.WriteFile("token.json", b, 0644)
}

func readFileJson(da *DataAuth) {

  file, err := ioutil.ReadFile("token.json")
  if err != nil {
    fmt.Println(err)
  }

  err = json.Unmarshal(file, &da)
	if err != nil {
		fmt.Println(err)
	}
}
