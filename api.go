package main

import (
	"cloud.google.com/go/firestore"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type ParamsInfo struct {
	Key      string
	Required bool
}

//API USER

func ApiUserExist(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	set := []ParamsInfo{
		{Key: "email", Required: true},
	}

	params, err := getParams(set, r)
	if err != nil {
		printLog("General", ip, "ApiUserExist", "Error to get params: "+err.Error(), 1)
		printErr(w, err.Error())
		return
	}

	exist := existUser(params["email"].(string))

	w.Write([]byte(fmt.Sprintf(`{"exist": "%v"}`, exist)))
}

//API AUTH

func ApiLogin(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	var params struct {
		Email    string
		Password string
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Verify password and email
	isValid, err := verifyAuth(params.Email, params.Password)
	if err != nil {
		if err.Error() == "inactive" {
			http.Error(w, `{"code": 401, "msg": "Account unactivated!"}`, http.StatusUnauthorized)
			return
		}

		//Nel caso in cui non è corretto perchè non c'è nel database
		printErr(w, "Incorrect username or password")
		return
	}

	if isValid == false {
		//Nel caso in cui c'è nel database ma non è corretta la password
		printErr(w, "Incorrect username or password")
		return
	}

	//Generate tokens
	tokens, err := GenerateTokenPair(params.Email)
	if err != nil {
		printLog("General", ip, "ApiLogin", "Error to generate token pair: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Set Cookies
	err = setCookies(w, tokens["RefreshToken"].Token)
	if err != nil {
		printLog("General", ip, "ApiLogin", "Error to set cookies: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Save refresh token in the database
	data := tokens["RefreshToken"].Fields
	exp := data["exp"].(time.Time).Unix()

	rf, ok := data["refreshId"].(string)
	if !ok {
		printLog("General", ip, "ApiLogin", "Error, the field with \"refreshId\" key doesn't exist", 1)
		printInternalErr(w)
		return
	}

	_, err = kaoriUser.Client.c.Collection("User").Doc(params.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUser.Client.ctx, map[string]int64{"exp": exp})

	if err != nil {
		printLog("General", ip, "ApiLogin", "Database operation error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Create JSON
	fields := tokens["AccessToken"].Fields
	exp, ok = fields["exp"].(int64)
	if !ok {
		printLog("General", ip, "ApiLogin", "Error, the field with \"AccessToken\" key doesn't exist", 1)
		printInternalErr(w)
		return
	}

	data2, err := json.Marshal(map[string]string{
		"AccessToken": tokens["AccessToken"].Token,
		"Expiration":  fmt.Sprint(exp),
	})
	if err != nil {
		printLog("General", ip, "ApiLogin", "Error to create AccessToken JSON: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	w.Write(data2)
}

func ApiRefresh(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	//Get token from cookies
	cookieData, err := getCookies(r)
	if err != nil {
		printLog("General", ip, "ApiRefresh", "Error to get cookies: "+err.Error(), 1)
		printErr(w, "")
		return
	}

	token := cookieData["RefreshToken"]

	//Extract data from token
	data, err := ExtractRefreshTokenMetadata(token, cfg.Password.RefreshToken)
	if err != nil {
		printLog("General", ip, "ApiRefresh", "Extract refresh token error: "+err.Error(), 1)
		printErr(w, "Cookies Not valid")
		return
	}

	//Check validity
	if !VerifyRefreshToken(data.Email, data.RefreshId) {
		printLog(data.Email, ip, "ApiRefresh", "Token not valid", 2)
		printErr(w, "Token not valid")
		return
	}

	if !VerifyExpireDate(data.Exp) {
		printLog(data.Email, ip, "ApiRefresh", "Token expired", 2)
		http.Error(w, `{"code": 401, "msg": "Token expired! Login required!"}`, http.StatusUnauthorized)
		return
	}

	//Remove old refresh token
	_, err = kaoriUser.Client.c.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(data.RefreshId).
		Delete(kaoriUser.Client.ctx)

	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Database connection error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Check token expired
	err = CheckOldsToken(data.Email)
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Check old tokens error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Generate new token pair
	tokens, err := GenerateTokenPair(data.Email)
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Generate token pair error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Set Cookies
	err = setCookies(w, tokens["RefreshToken"].Token)
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Set cookie error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Save refresh token in the database
	data2 := tokens["RefreshToken"].Fields
	exp := data2["exp"].(time.Time).Unix()

	rf, ok := data2["refreshId"].(string)
	if !ok {
		printLog(data.Email, ip, "ApiRefresh", `The field with key "refreshId" doesn't exist'`, 1)
		printInternalErr(w)
		return
	}

	_, err = kaoriUser.Client.c.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUser.Client.ctx, map[string]int64{"exp": exp})

	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Database connection error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Create JSON
	fields := tokens["AccessToken"].Fields
	exp, ok = fields["exp"].(int64)
	if !ok {
		printLog(data.Email, ip, "ApiRefresh", `The field with key "exp" doesn't exist'`, 1)
		printInternalErr(w)
		return
	}

	//Create JSON
	jsonData, err := json.Marshal(map[string]string{
		"AccessTocken": tokens["AccessToken"].Token,
		"Expiration":   fmt.Sprint(exp),
	})
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Create JSON error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	w.Write(jsonData)
}

func ApiSignUp(w http.ResponseWriter, r *http.Request) {

	// Declare a new MusicData struct.
	var u User

	ip := GetIP(r)

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		printErr(w, err.Error())
		return
	}

	err = u.IsValid()
	if err != nil {
		printErr(w, err.Error())
		return
	}

	if existUser(u.Email) {
		printErr(w, "A user with this email already exists")
		return
	}

	u.NewUser() //Set default value

	//Add in the database
	err = u.AddNewUser()
	if err != nil {
		printLog("General", ip, "ApiSignUp", "AddNewUser error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Generate tokens
	tokens, err := GenerateTokenPair(u.Email)
	if err != nil {
		printLog("General", ip, "ApiSignUp", "Error to generate token pair: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	type Conferma struct {
		Username string
		Link     string
		Login    string
	}

	//Save refresh token in the database
	data := tokens["RefreshToken"].Fields
	exp := data["exp"].(time.Time).Unix()

	rf, ok := data["refreshId"].(string)
	if !ok {
		printLog("General", ip, "ApiSignUp", `The field with key "refreshId" doesn't exist'`, 1)
		printInternalErr(w)
		return
	}

	_, err = kaoriUser.Client.c.Collection("User").Doc(u.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUser.Client.ctx, map[string]int64{"exp": exp})

	if err != nil {
		printLog("General", ip, "ApiSignUp", "Database error: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	c := Conferma{
		Username: u.Username,
		Link: fmt.Sprintf(
			"https://%s%sconfirm?email=%s&id=%s",
			cfg.Server.Host+cfg.Server.Port,
			endpointAuth,
			u.Email,
			rf,
		),
		Login: "https://" + cfg.Server.Host + cfg.Server.Port + endpointLogin.String(),
	}

	//Send mails
	signupField := cfg.Template.Mail.Fields["registration"]
	err = sendEmail(u.Email, signupField.Object, c, filepath.Join(cfg.Template.Mail.Path, signupField.File))
	if err != nil {
		printLog("General", ip, "ApiSignUp", "Error to send mail: "+err.Error(), 1)
		return
	}

}

func ApiConfirmSignUp(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	params := []ParamsInfo{
		{Key: "id", Required: true},
		{Key: "email", Required: true},
	}

	p, err := getParams(params, r)
	if err != nil {
		printLog("General", ip, "ApiConfirmSignup", "Error to get params: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	if !VerifyRefreshToken(p["email"].(string), p["id"].(string)) {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Warning to verify refresh token: Token not valid", 2)
		printErr(w, "Token not valid!")
		return
	}

	//Remove old refresh token
	_, err = kaoriUser.Client.c.Collection("User").Doc(p["email"].(string)).
		Collection("RefreshToken").Doc(p["id"].(string)).
		Delete(kaoriUser.Client.ctx)

	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Set user to active
	_, err = kaoriUser.Client.c.Collection("User").Doc(p["email"].(string)).
		Set(kaoriUser.Client.ctx, map[string]bool{"IsActive": true}, firestore.MergeAll)

	//Redirect to login
	data, err := parseTemplateHtml(cfg.Template.Html.Redirect, "https://"+cfg.Server.Host+cfg.Server.Port+endpointLogin.String())
	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	w.Write([]byte(data))
}

//API ADD-DATA

func ApiAddMusic(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	//Extract token JWT
	tokenString := ExtractToken(r)
	metadata, err := ExtractAccessTokenMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		printErr(w, err.Error())
		return
	}

	// Declare a new MusicData struct.
	var md MusicData

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&md)
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Warning to decode JSON: "+err.Error(), 2)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(md.IdAnilist)

	//TODO: Forse potrei controllare anche che non ci siano virus
	err = md.CheckError()
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Warning to check input users: "+err.Error(), 2)
		printErr(w, err.Error())
		return
	}

	//Check if it already exist
	_, err = kaoriTmp.Client.c.Collection(md.Type).
		Doc(strconv.Itoa(md.IdAnilist)).
		Get(kaoriTmp.Client.ctx)
	if err == nil {
		printLog(metadata.Email, ip, "ApiAddMusic", fmt.Sprintf("Warning item with id=%s already exist: ", md.IdAnilist), 2)
		http.Error(w, `{"code": 409, "msg": "The track already exists"}`, http.StatusConflict)
		return
	}

	md.GetNameAnime()
	err = md.NormalizeName()
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Error to normalize name: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Upload
	err = md.UploadTemporaryFile()
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Error to upload temporary file: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Add to database
	err = md.AddDataToTmpDatabase()
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Error to add music data to the temp database: "+err.Error(), 1)
		printInternalErr(w)
		return
	}
}
