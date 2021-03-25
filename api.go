package main

import (
	"cloud.google.com/go/firestore"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ParamsInfo struct {
	Key      string
	Required bool
}

//API USER

func ApiUserExist(w http.ResponseWriter, r *http.Request) {

	var exist bool

	set := []ParamsInfo{
		{Key: "email", Required: true},
	}

	params, err := getParams(set, r)
	if err != nil {
		log.Println(err)
		printErr(w, err.Error())
		return
	}
	_, err = kaoriUser.Client.GetItem("User", params["email"].(string))
	if err != nil {
		exist = false
	} else {
		exist = true
	}

	w.Write([]byte(fmt.Sprintf(`{"exist": "%v"}`, exist)))
}

//API AUTH

func ApiLogin(w http.ResponseWriter, r *http.Request) {

	var params struct {
		Email string
		Password string
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Verify password and email
	isValid, err := verifyAuth(params.Email, params.Password)

	if err.Error() == "unactive" {
		w.WriteHeader(401)
		w.Write([]byte("Account inattivo!"))
		return
	}

	if err != nil  || isValid == false {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Nome utente e/o password errati."))
		return
	}

	//Generate tokens
	tokens, err := GenerateTokenPair(params.Email)
	if err != nil {
		printInternalErr(w)
		return
	}

	//Set Cookies
	err = setCookies(w, tokens["RefreshToken"].Token)
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Save refresh token in the database
	data := tokens["RefreshToken"].Fields

	exp := data["exp"].(time.Time).Unix()

	rf, ok := data["refreshId"].(string)
	if !ok {
		printInternalErr(w)
		return
	}

	_, err = kaoriUser.Client.c.Collection("User").Doc(params.Email).
								Collection("RefreshToken").Doc(rf).
								Set(kaoriUser.Client.ctx, map[string]int64{"exp": exp})

	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Create JSON
	fields := tokens["AccessToken"].Fields
	exp, ok = fields["exp"].(int64)
	if !ok {
		printInternalErr(w)
		return
	}

	data2, err := json.Marshal(map[string]string{
		"AccessToken": tokens["AccessToken"].Token,
		"Expiration": fmt.Sprint(exp),
	})
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	w.Write(data2)
}

func ApiRefresh(w http.ResponseWriter, r *http.Request){

	//Get token from cookies
	cookieData, err := getCookies(r)
	if err != nil {
		log.Println(err)
		printErr(w, "")
		return
	}

	token := cookieData["RefreshToken"]

	//Extract data from token
	data, err := ExtractRefreshTokenMetadata(token, os.Getenv("REFRESH_SECRET"))
	if err != nil {
		printErr(w, err.Error())
		return
	}

	//Check validity

	if !VerifyRefreshToken(data.Email, data.RefreshId) {
		printErr(w, errors.New("Token not valid").Error())
		return
	}

	if !VerifyExpireDate(data.Exp){
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token expired! Login required!"))
		return
	}

	//Remove old refresh token
	_, err = kaoriUser.Client.c.Collection("User").Doc(data.Email).
								 Collection("RefreshToken").Doc(data.RefreshId).
								 Delete(kaoriUser.Client.ctx)

	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Check token expired
	err = CheckOldsToken(data.Email)
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}


	//Generate new token pair
	tokens, err := GenerateTokenPair(data.Email)
	if err != nil {
		printInternalErr(w)
		return
	}

	//Set Cookies
	err = setCookies(w, tokens["RefreshToken"].Token)
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Save refresh token in the database
	data2 := tokens["RefreshToken"].Fields

	exp := data2["exp"].(time.Time).Unix()

	rf, ok := data2["refreshId"].(string)
	if !ok {
		printInternalErr(w)
		return
	}

	_, err = kaoriUser.Client.c.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUser.Client.ctx, map[string]int64{"exp": exp})

	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Create JSON
	fields := tokens["AccessToken"].Fields
	exp, ok = fields["exp"].(int64)
	if !ok {
		printInternalErr(w)
		return
	}

	//Create JSON
	jsonData, err := json.Marshal(map[string]string{
		"AccessTocken": tokens["AccessToken"].Token,
		"Expiration": fmt.Sprint(exp),
	})
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	w.Write(jsonData)
}

func ApiSignUp(w http.ResponseWriter, r *http.Request) {

	// Declare a new MusicData struct.
	var u User

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//TODO: Verifica che i dati inviati siano corretti.
	//TODO: Verifica che l'utente non esista (quindi che non ci sia la mail).

	u.NewUser() //Set default value

	//Add in the database
	err = u.AddNewUser()
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Generate tokens
	tokens, err := GenerateTokenPair(u.Email)
	if err != nil {
		printInternalErr(w)
		return
	}

	type Conferma struct {
		Username string
		Link string
		Login string
	}

	//Save refresh token in the database
	data := tokens["RefreshToken"].Fields

	exp := data["exp"].(time.Time).Unix()

	rf, ok := data["refreshId"].(string)
	if !ok {
		printInternalErr(w)
		return
	}

	_, err = kaoriUser.Client.c.Collection("User").Doc(u.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUser.Client.ctx, map[string]int64{"exp": exp})

	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//TODO: Add base_url
	c := Conferma{
		Username: u.Username,
		Link: filepath.Join(baseUrl, fmt.Sprintf(  "/api/auth/confirm?email=%s&id=%s", u.Email, rf)),
		Login: filepath.Join(baseUrl, "/KaoriGui/login.html"),
	}


	//Send mails
	err = sendEmail(u.Email, "CONFERMA REGISTRAZIONE", c, filepath.Join(emailTemplate, "registrazione.txt"))
	if err != nil {
		log.Println(err)
		return
	}
}

func ApiConfirmSignUp(w http.ResponseWriter, r *http.Request) {

	params := []ParamsInfo{
		{Key:      "id", Required: true},
		{Key:      "email", Required: true},
	}

	p, err := getParams(params, r)
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	if !VerifyRefreshToken(p["email"].(string), p["id"].(string)) {
		printErr(w, errors.New("Token not valid").Error())
		return
	}

	//Remove old refresh token
	_, err = kaoriUser.Client.c.Collection("User").Doc(p["email"].(string)).
		Collection("RefreshToken").Doc(p["id"].(string)).
		Delete(kaoriUser.Client.ctx)

	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Set user to active
	_, err = kaoriUser.Client.c.Collection("User").Doc(p["email"].(string)).
								Set(kaoriUser.Client.ctx, map[string]bool{"IsActive": true}, firestore.MergeAll)

	content, err := os.ReadFile("KaoriGui/login.html")
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	w.Write(content)
}

//API ADD-DATA

func ApiAddMusic(w http.ResponseWriter, r *http.Request) {

	//TODO: Verificare che l'utente (se ha solo i permessi di utente) non abbia già caricato il massimo di canzoni (cioè 10)

	// Declare a new MusicData struct.
	var md MusicData

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&md)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//TODO: Check size files
	//TODO: Forse potrei controllare anche che non ci siano virus
	err = md.CheckError()
	if err != nil {
		log.Println(err)
		printErr(w, err.Error())
		return
	}

	md.GetNameAnime()
	fmt.Println(md.AnimeName)
	err = md.NormalizeName()
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}

	//Upload
	err = md.UploadTemporaryFile()
	if err != nil {
		printInternalErr(w)
		return
	}

	//TODO: Verificare che non sia già nel database

	//Add to database
	err = md.AddDataToTmpDatabase()
	if err != nil {
		log.Println(err)
		printInternalErr(w)
		return
	}
}
