package main

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/firestore"
	"encoding/json"
	"fmt"
	kaoriAnime "github.com/CodeOfTheKnight/Kaori/kaoriData/anime"
	kaoriManga "github.com/CodeOfTheKnight/Kaori/kaoriData/manga"
	kaoriMusic "github.com/CodeOfTheKnight/Kaori/kaoriData/music"
	kaoriUser "github.com/CodeOfTheKnight/Kaori/kaoriData/user"
	"github.com/CodeOfTheKnight/Kaori/kaoriJwt"
	"github.com/CodeOfTheKnight/Kaori/kaoriLog"
	"github.com/CodeOfTheKnight/Kaori/kaoriMail"
	"github.com/CodeOfTheKnight/Kaori/kaoriSettings"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	anilist "github.com/kaiserbh/anilistgo/anilist/query"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//API USER

func ApiUserInfo(w http.ResponseWriter, r *http.Request){

	var u kaoriUser.User

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	//Get document
	document, err := kaoriUserDB.Client.C.Collection("User").
										Doc(mappa.Get("email")).
										Get(kaoriUserDB.Client.Ctx)

	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	err = document.DataTo(&u)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error to create user struct: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	data, err := json.Marshal(u)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error create JSON: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)
}

//API USER SETTINGS

func ApiSettingsGet(w http.ResponseWriter, r *http.Request) {

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "[Start] Get settings", 0)

	//Get document
	document, err := kaoriUserDB.Client.C.Collection("User").
		Doc(mappa.Get("email")).
		Get(kaoriUserDB.Client.Ctx)

	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	data, err := json.Marshal(document.Data()["Settings"])
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "Error create JSON: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "[Done] Get settings", 0)

	w.Write(data)

}

func ApiSettingsSet(w http.ResponseWriter, r *http.Request){

	//var s Settings
	var u kaoriUser.User
	var m map[string]interface{}

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "[Start] Change settings", 0)

	//Read precedent settings
	document, err := kaoriUserDB.Client.C.Collection("User").
										Doc(mappa.Get("email")).
										Get(kaoriUserDB.Client.Ctx)

	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Parse document firebase in users struct
	err = document.DataTo(&u)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to conversion in settings: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Merge with the form submitted by users.
	err = json.NewDecoder(r.Body).Decode(&u.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Conversion struct settings in a map
	err = mapstructure.Decode(u.Settings, &m)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to conversion in a map: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	fmt.Println("MAPPA:", m)

	fmt.Println("Settings:", u.Settings)

	//Check settings value
	if err = u.Settings.IsValid(); err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Data is invalid: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	//Send map to the database
	_, err =  kaoriUserDB.Client.C.Collection("User").
		Doc(mappa.Get("email")).
		Set(kaoriUserDB.Client.Ctx, map[string]interface{}{"Settings": m}, firestore.MergeAll)

	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Make JSON response
	data, err := json.Marshal(m)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to create JSON: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "[Done] Change settings", 0)

	w.Write(data)
}

//API AUTH

func ApiUserExist(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	set := []kaoriUtils.ParamsInfo{
		{Key: "email", Required: true},
	}

	params, err := kaoriUtils.GetParams(set, r)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiUserExist", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	exist := kaoriUtils.ExistUser(kaoriUserDB, params["email"].(string))

	w.Write([]byte(fmt.Sprintf(`{"exist": "%v"}`, exist)))
}

func ApiLogin(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

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
	isValid, err := kaoriUtils.VerifyAuth(kaoriUserDB, params.Email, params.Password)
	if err != nil {
		if err.Error() == "inactive" {
			http.Error(w, `{"code": 401, "msg": "Account unactivated!"}`, http.StatusUnauthorized)
			return
		}

		//Nel caso in cui non è corretto perchè non c'è nel database
		kaoriUtils.PrintErr(w, "Incorrect username or password")
		return
	}

	if isValid == false {
		//Nel caso in cui c'è nel database ma non è corretta la password
		kaoriUtils.PrintErr(w, "Incorrect username or password")
		return
	}

	//Generate tokens
	perm, err := kaoriJwt.GetPermissionFromDB(kaoriUserDB, params.Email)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to get user permission from db: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	tokenPair := kaoriJwt.NewJWTokenPair(cfg.Jwt.Exp.AccessToken, cfg.Jwt.Exp.RefreshToken)
	tokenPair.Access.Obj.Email = params.Email
	tokenPair.Access.Obj.Iss = cfg.Jwt.Iss
	tokenPair.Access.Obj.Iat = time.Now().Unix()
	tokenPair.Access.Obj.Company = cfg.Jwt.Company
	tokenPair.Access.Obj.Permission = perm.ToString()
	tokenPair.Refresh.Obj.Email = params.Email
	tokenPair.Refresh.Obj.RefreshId = kaoriUtils.GenerateID()

	err = tokenPair.GenerateTokenPair(cfg.Password.AccessToken, cfg.Password.RefreshToken)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to generate token pair: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Set Cookies
	err = kaoriUtils.SetCookies(w, tokenPair.Refresh.Token, cfg.Jwt.Iss, cfg.Password.Cookies)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to set cookies: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	_, err = kaoriUserDB.Client.C.Collection("User").Doc(params.Email).
		Collection("RefreshToken").Doc(tokenPair.Refresh.Obj.RefreshId).
		Set(kaoriUserDB.Client.Ctx, map[string]int64{"exp": tokenPair.Refresh.Obj.Exp})

	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Database operation error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	data2, err := json.Marshal(map[string]string{
		"AccessToken": tokenPair.Access.Token,
		"Expiration":  fmt.Sprint(tokenPair.Access.Obj.Exp),
	})
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to create AccessToken JSON: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data2)
}

func ApiRefresh(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	//Get token from cookies
	cookieData, err := kaoriUtils.GetCookies(r, cfg.Jwt.Iss, cfg.Password.Cookies)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiRefresh", "Error to get cookies: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "")
		return
	}

	token := cookieData["RefreshToken"]

	//Extract data from token
	data, err := kaoriJwt.ExtractRefreshMetadata(token, cfg.Password.RefreshToken)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiRefresh", "Extract refresh token error: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Cookies Not valid")
		return
	}

	//Check validity
	if !kaoriJwt.VerifyRefreshToken(kaoriUserDB, data.Email, data.RefreshId) {
		kaoriLog.PrintLog(data.Email, ip, "ApiRefresh", "Token not valid", 2)
		kaoriUtils.PrintErr(w, "Token not valid")
		return
	}

	if !kaoriJwt.VerifyExpireDate(data.Exp) {
		kaoriLog.PrintLog(data.Email, ip, "ApiRefresh", "Token expired", 2)
		http.Error(w, `{"code": 401, "msg": "Token expired! Login required!"}`, http.StatusUnauthorized)
		return
	}

	//Remove old refresh token
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(data.RefreshId).
		Delete(kaoriUserDB.Client.Ctx)

	if err != nil {
		kaoriLog.PrintLog(data.Email, ip, "ApiRefresh", "Database connection error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Check token expired
	err = kaoriJwt.CheckOldsToken(kaoriUserDB, data.Email)
	if err != nil {
		kaoriLog.PrintLog(data.Email, ip, "ApiRefresh", "Check old tokens error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Generate new token pair
	perm, err := kaoriJwt.GetPermissionFromDB(kaoriUserDB, data.Email)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to get user permission from db: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	tokenPair := kaoriJwt.NewJWTokenPair(cfg.Jwt.Exp.AccessToken, cfg.Jwt.Exp.RefreshToken)
	tokenPair.Access.Obj.Email = data.Email
	tokenPair.Access.Obj.Iss = cfg.Jwt.Iss
	tokenPair.Access.Obj.Iat = time.Now().Unix()
	tokenPair.Access.Obj.Company = cfg.Jwt.Company
	tokenPair.Access.Obj.Permission = perm.ToString()
	tokenPair.Refresh.Obj.Email = data.Email
	tokenPair.Refresh.Obj.RefreshId = kaoriUtils.GenerateID()

	err = tokenPair.GenerateTokenPair(cfg.Password.AccessToken, cfg.Password.RefreshToken)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to generate token pair: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Set Cookies
	err = kaoriUtils.SetCookies(w, tokenPair.Refresh.Token, cfg.Jwt.Iss, cfg.Password.Cookies)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to set cookies: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	_, err = kaoriUserDB.Client.C.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(tokenPair.Refresh.Obj.RefreshId).
		Set(kaoriUserDB.Client.Ctx, map[string]int64{"exp": tokenPair.Refresh.Obj.Exp})

	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Database operation error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	data2, err := json.Marshal(map[string]string{
		"AccessToken": tokenPair.Access.Token,
		"Expiration":  fmt.Sprint(tokenPair.Access.Obj.Exp),
	})
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiLogin", "Error to create AccessToken JSON: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data2)
}

func ApiSignUp(w http.ResponseWriter, r *http.Request) {

	var u kaoriUser.User

	// Declare a new MusicData struct.
	user := struct {
		Email          string `json:"email"`
		Username       string `json:"username"`
		Password       string `json:"password"`
		ProfilePicture string `json:"profilePicture,omitempty"`
	}{}

	ip := kaoriUtils.GetIP(r)

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	u.Email = user.Email
	u.Username = user.Username
	u.Password = user.Password
	u.ProfilePicture = user.ProfilePicture

	err = u.IsValid()
	if err != nil {
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	if kaoriUtils.ExistUser(kaoriUserDB, u.Email) {
		kaoriUtils.PrintErr(w, "A user with this email already exists")
		return
	}

	u.NewUser() //Set default value

	//Add in the database
	err = u.AddNewUser(kaoriUserDB)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiSignUp", "AddNewUser error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Generate refresh token
	refToken := kaoriJwt.NewRefreshToken(cfg.Jwt.Exp.RefreshToken)
	refToken.RefreshId = kaoriUtils.GenerateID()
	refToken.Email = u.Email

	/*
	_, err := refToken.GenerateToken(cfg.Password.RefreshToken)
	if err != nil {
		printLog("General", ip, "ApiSignUp", "Error to generate token pair: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
	*/

	type Conferma struct {
		Username string
		LinkConfirm     string
		LinkReject string
	}

	//Save refresh token in the database
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(u.Email).
		Collection("RefreshToken").Doc(refToken.RefreshId).
		Set(kaoriUserDB.Client.Ctx, map[string]int64{"exp": refToken.Exp})

	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiSignUp", "Database error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	c := Conferma{
		Username: u.Username,
		LinkConfirm: fmt.Sprintf(
			"https://%s%sconfirm?email=%s&id=%s",
			cfg.Server.Host+cfg.Server.Port,
			endpointAuth,
			u.Email,
			refToken.RefreshId,
		),
		LinkReject: fmt.Sprintf(
			"https://%s%sreject?email=%s&id=%s",
			cfg.Server.Host+cfg.Server.Port,
			endpointAuth,
			u.Email,
			refToken.RefreshId,
		),
	}

	//Send mails
	signupField := cfg.Template.Mail["registration"]
	err = kaoriMail.SendEmail(
		cfg.Mail.SmtpServer.Host + cfg.Mail.SmtpServer.Port,
		cfg.Mail.Address,
		cfg.Password.Mail,
		u.Email,
		signupField.Object,
		signupField.File,
		c,
	)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiSignUp", "Error to send mail: "+err.Error(), 1)
		return
	}

}

func ApiConfirmSignUp(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	params := []kaoriUtils.ParamsInfo{
		{Key: "id", Required: true},
		{Key: "email", Required: true},
	}

	p, err := kaoriUtils.GetParams(params, r)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiConfirmSignup", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	if !kaoriJwt.VerifyRefreshToken(kaoriUserDB, p["email"].(string), p["id"].(string)) {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Warning to verify refresh token: Token not valid", 2)
		kaoriUtils.PrintErr(w, "Token not valid!")
		return
	}

	//Remove old refresh token
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(p["email"].(string)).
		Collection("RefreshToken").Doc(p["id"].(string)).
		Delete(kaoriUserDB.Client.Ctx)

	if err != nil {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Set user to active
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(p["email"].(string)).
		Set(kaoriUserDB.Client.Ctx, map[string]bool{"IsActive": true}, firestore.MergeAll)

	if err != nil {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Redirect to login
	data, err := kaoriUtils.ParseTemplateHtml(cfg.Template.Html["redirect"], "https://"+cfg.Server.Host+cfg.Server.Port+endpointLogin.String())
	if err != nil {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Error to create template: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write([]byte(data))
}

func ApiRejectSignUp(w http.ResponseWriter, r *http.Request){

	ip := kaoriUtils.GetIP(r)

	params := []kaoriUtils.ParamsInfo{
		{Key: "id", Required: true},
		{Key: "email", Required: true},
	}

	p, err := kaoriUtils.GetParams(params, r)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiConfirmSignup", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	if !kaoriJwt.VerifyRefreshToken(kaoriUserDB, p["email"].(string), p["id"].(string)) {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Warning to verify refresh token: Token not valid", 2)
		kaoriUtils.PrintErr(w, "Token not valid!")
		return
	}

	//Remove old refresh token
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(p["email"].(string)).
		Collection("RefreshToken").Doc(p["id"].(string)).
		Delete(kaoriUserDB.Client.Ctx)

	if err != nil {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Remove user
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(p["email"].(string)).Delete(kaoriUserDB.Client.Ctx)
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
	}

	if err != nil {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Redirect to login
	data, err := kaoriUtils.ParseTemplateHtml(cfg.Template.Html["redirect"], "https://"+cfg.Server.Host+cfg.Server.Port+endpointLogin.String())
	if err != nil {
		kaoriLog.PrintLog(p["email"].(string), ip, "ApiConfirmSignup", "Error to create template: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write([]byte(data))
}

//API SERVICE

func ApiServiceAnime(w http.ResponseWriter, r *http.Request) {

	type data struct {
        Id string
		Info anilist.Media
		Data kaoriAnime.Anime
	}
	var d data

	//var a kaoriAnime.Anime
	var err error

	//GetIP
	ip := kaoriUtils.GetIP(r)
	
	//Get anime id
	id := filepath.Base(r.URL.Path)
	idNum, err := strconv.Atoi(id)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ServiceAnime", err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	fmt.Println("ID:", idNum)


	//Get information from anilist
	m := anilist.Media{}
	err = m.FilterAnimeByID(idNum)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceAnime", "Serving anime error: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id not found")
		return
	}

	//Get video info
	d.Data.Episodes, err = kaoriAnime.GetEpisodesFromDB(kaoriAnimeDB.Client, idNum)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceAnime", "Error to get anime episodes: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id not found")
		return
	}
	d.Info = m
	d.Id = id

	respData, err := kaoriUtils.ParseTemplateHtml("kaoriSrc/template/html/anime/anime.html", d)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceAnime", "Error to parse template: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write([]byte(respData))
}

func ApiServiceEpisode(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	params := []kaoriUtils.ParamsInfo{
		{Key: "id", Required: true},
		{Key: "num", Required: true},
	}

	p, err := kaoriUtils.GetParams(params, r)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceEpisode", "Error to get episode: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Params not valid")
		return
	}

	idNum, err := strconv.Atoi(p["id"].(string))
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceEpisode", err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id or num page not valid")
		return
	}

	pageNum, err := strconv.Atoi(p["num"].(string))
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceEpisode", err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id or num page not valid")
		return
	}

	fmt.Println(idNum, pageNum)

	episode, err := kaoriAnime.GetEpisodeFromDB(kaoriAnimeDB.Client, idNum, pageNum)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceEpisode", "Error to get episode: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id or episode not found.")
		return
	}

	data, err := json.Marshal(episode.Videos)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceEpisode", "Error to get episode: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)
}

func ApiServiceManga(w http.ResponseWriter, r *http.Request) {

	type data struct {
		Info anilist.Media
		Data kaoriManga.Manga
		Conf kaoriSettings.ServerConfig
	}
	var d data

	//var a kaoriAnime.Anime
	var err error

	//GetIP
	ip := kaoriUtils.GetIP(r)

	//Get anime id
	id := filepath.Base(r.URL.Path)
	idNum, err := strconv.Atoi(id)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ServiceManga", err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	fmt.Println("ID:", idNum)


	//Get information from anilist
	m := anilist.Media{}
	err = m.FilterAnimeByID(idNum)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceManga", "Serving manga error: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id not found")
		return
	}

	//Get video info
	d.Data.Chapters, err = kaoriManga.GetChaptersFromDB(kaoriAnimeDB.Client, idNum)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceManga", "Error to get manga chapter: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id not found")
		return
	}
	d.Info = m

	respData, err := kaoriUtils.ParseTemplateHtml("kaoriSrc/template/html/manga/manga.html", d)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceManga", "Error to parse template: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write([]byte(respData))

}

func ApiServiceChapter(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	params := []kaoriUtils.ParamsInfo{
		{Key: "id", Required: true},
		{Key: "num", Required: true},
	}

	p, err := kaoriUtils.GetParams(params, r)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceChapter", "Error to get chapter: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, "Params not valid")
		return
	}

	idNum, err := strconv.Atoi(p["id"].(string))
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceChapter", err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id or num page not valid")
		return
	}

	chNum, err := strconv.Atoi(p["num"].(string))
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceChapter", err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id or num page not valid")
		return
	}

	fmt.Println(idNum, chNum)

	chapter, err := kaoriManga.GetChapterFromDB(kaoriAnimeDB.Client, idNum, chNum)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceEpisode", "Error to get episode: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Id or episode not found.")
		return
	}

	data, err := json.Marshal(chapter.Pages)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiServiceChapter", "Error to get episode: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)

}

//API ADD-DATA

func ApiAddMusic(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	//Extract token JWT
	tokenString := kaoriJwt.ExtractToken(r)
	metadata, err := kaoriJwt.ExtractAccessMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiConfigGet", "Error to extract access token metadata: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	// Declare a new MusicData struct.
	var md kaoriMusic.MusicData

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&md)
	if err != nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiAddMusic", "Warning to decode JSON: "+err.Error(), 2)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(md.IdAnilist)

	//TODO: Forse potrei controllare anche che non ci siano virus
	err = md.CheckError()
	if err != nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiAddMusic", "Warning to check input users: "+err.Error(), 2)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	//Check if it already exist
	_, err = kaoriTmpDB.Client.C.Collection(md.Type).
		Doc(strconv.Itoa(md.IdAnilist)).
		Get(kaoriTmpDB.Client.Ctx)
	if err == nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiAddMusic", fmt.Sprintf("Warning item with id=%s already exist: ", md.IdAnilist), 2)
		http.Error(w, `{"code": 409, "msg": "The track already exists"}`, http.StatusConflict)
		return
	}

	md.GetNameAnime()
	err = md.NormalizeName(cfg.Template.Music["name"])
	if err != nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiAddMusic", "Error to normalize name: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Upload
	err = md.UploadTemporaryFile()
	if err != nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiAddMusic", "Error to upload temporary file: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Add to database
	err = md.AddDataToTmpDatabase(kaoriTmpDB)
	if err != nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiAddMusic", "Error to add music data to the temp database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
}

//API ADMIN

func ApiConfigGet(w http.ResponseWriter, r *http.Request){

	ip := kaoriUtils.GetIP(r)

	//Extract token JWT
	tokenString := kaoriJwt.ExtractToken(r)
	metadata, err := kaoriJwt.ExtractAccessMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		kaoriLog.PrintLog("General", ip, "ApiConfigGet", "Error to extract access token metadata: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		kaoriLog.PrintLog(metadata.Email, ip, "ApiConfigGet", "Error to create JSON of config struct: " + err.Error(), 1)
	}

	w.Write(data)

	kaoriLog.PrintLog(metadata.Email, ip, "ApiConfigGet", "The user has read the settings", 0)
}

func ApiConfigSet(w http.ResponseWriter, r *http.Request){

	var cfg2 kaoriSettings.Config
	cfg2 = *cfg

	mappa := r.Context().Value("values").(ContextValues)
	data, _ := io.ReadAll(r.Body)

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&cfg2)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Error to create JSON: " + err.Error(), 1)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//TODO: add
	/*
	//Check configuration validity
	if err = cfg2.CheckConfig(); err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Config not valid: " + err.Error(), 1)
		printErr(w, "Config not valid: " + err.Error())
		return
	}
	 */

	wdone, err := cfg2.WriteConfig()
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Error to write config: "+err.Error(), 1)
		http.Error(w, fmt.Sprintf("{\"code\": 500, \"msg\": \"%s\"}\n", strings.Join(wdone, "\n")), http.StatusInternalServerError)
		return
	}

	//TODO: Change file config
	//TODO: Reboot
	fmt.Println(cfg2)

	w.Write([]byte(fmt.Sprintf("{\"code\": 200, \"msg\": \"\n%s\"}\n", strings.Join(wdone, "\n"))))
}

func ApiLogServer(w http.ResponseWriter, r *http.Request){

	var logs []*kaoriLog.ServerLog
	var err error
	var params map[string]interface{}
	var logsString []string

	mappa := r.Context().Value("values").(ContextValues)


	if r.Method == http.MethodGet{
		set := []kaoriUtils.ParamsInfo{
			{Key: "func", Required: false},
			{Key: "ip", Required: false},
			{Key: "user", Required: false},
			{Key: "msg", Required: false},
			{Key: "level", Required: false},
			{Key: "time", Required: false},
			{Key: "order", Required: true},
		}

		params, err = kaoriUtils.GetParams(set, r)
		if err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to get params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogGet(set, params); err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

	} else {

		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to get JSON params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogPost(params); err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

	}


	f, err := os.Open("log/server.log.json")
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to open log file: "+ err.Error(),1 )
		kaoriUtils.PrintInternalErr(w)
		return
	}

	orderSlice := strings.Split(params["order"].(string), ",")

	//Primo grande filtro

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		val := scanner.Bytes()

		var objmap map[string]json.RawMessage
		err = json.Unmarshal(val, &objmap)
		if err != nil {
			log.Fatal(err)
			return
		}

		str, err := objmap[orderSlice[0]].MarshalJSON()
		if err != nil {
			log.Fatal(err)
			return
		}

		if strings.Trim(string(str), "\"") == params[orderSlice[0]].(string) {
			logsString = append(logsString, string(val))
		}


	}

	if len(orderSlice) != 1 {

		//Other filters
		for _, filter := range orderSlice[1:] {
			logsString, err = kaoriUtils.FilterLog(logsString, filter, params[filter].(string))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	for _, item := range logsString {

		var sl kaoriLog.ServerLog

		err = json.Unmarshal([]byte(item), &sl)
		if err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of single log: "+ err.Error(),1 )
			kaoriUtils.PrintInternalErr(w)
			return
		}

		logs = append(logs, &sl)

	}

	data, err := json.Marshal(logs)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of slice log: "+ err.Error(),1 )
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)
}

func ApiLogConnection(w http.ResponseWriter, r *http.Request){

	var logsString []string
	var logs []*kaoriLog.HTTPReqInfo
	var err error
	var params map[string]interface{}

	mappa := r.Context().Value("values").(ContextValues)

	if r.Method == http.MethodGet {
		set := []kaoriUtils.ParamsInfo{
			{Key: "method", Required: false},
			{Key: "url", Required: false},
			{Key: "ref", Required: false},
			{Key: "ip", Required: false},
			{Key: "code", Required: false},
			{Key: "size", Required: false},
			{Key: "duration", Required: false},
			{Key: "data", Required: false},
			{Key: "agent", Required: false},
			{Key: "order", Required: true},
		}

		params, err = kaoriUtils.GetParams(set, r)
		if err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to get params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogGet(set, params); err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

	} else {
		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to get JSON params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogPost(params); err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}
	}


	f, err := os.Open("log/connection.log.json")
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to open log file: "+ err.Error(),1 )
		kaoriUtils.PrintInternalErr(w)
		return
	}

	orderSlice := strings.Split(params["order"].(string), ",")

	//Primo grande filtro

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		val := scanner.Bytes()

		var objmap map[string]json.RawMessage
		err = json.Unmarshal(val, &objmap)
		if err != nil {
			log.Fatal(err)
			return
		}

		str, err := objmap[orderSlice[0]].MarshalJSON()
		if err != nil {
			log.Fatal(err)
			return
		}

		if strings.Trim(string(str), "\"") == params[orderSlice[0]].(string) {
			logsString = append(logsString, string(val))
		}


	}

	if len(orderSlice) != 1 {

		//Other filters
		for _, filter := range orderSlice[1:] {
			logsString, err = kaoriUtils.FilterLog(logsString, filter, params[filter].(string))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}


	for _, item := range logsString {

		var hl kaoriLog.HTTPReqInfo

		err = json.Unmarshal([]byte(item), &hl)
		if err != nil {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to create JSON of single log: "+ err.Error(),1 )
			kaoriUtils.PrintInternalErr(w)
			return
		}

		logs = append(logs, &hl)

	}

	data, err := json.Marshal(logs)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of slice log: "+ err.Error(),1 )
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)

}

func ApiAnimeInsert(w http.ResponseWriter, r *http.Request) {

	var a kaoriAnime.Anime

	mappa := r.Context().Value("values").(ContextValues)

	idAnilist := filepath.Base(r.URL.Path)

	//Read client data
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	num, err := strconv.Atoi(idAnilist)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	if a.Id != 0 {

		if a.Id != num {
			kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiInsertAnime", "Error, id of body request and id in the URL don't match", 1)
			kaoriUtils.PrintErr(w, "Error, id of body request and id in the URL don't match")
			return
		}

	} else {
		a.Id = num
	}

	//NO RELATIONAL DATABASE
	/*
	err = a.SendToDb(kaoriDataDB)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}
	*/

	err = a.SendToDbRel(kaoriAnimeDB.Client)
	if err != nil {
		strError := fmt.Sprintf("Error to insert anime in the relational database [%d]: %s", a.Id, err.Error())
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("id"), "ApiAnimeInsert", strError , 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	kaoriUtils.PrintOk(w)
}

/*
func ApiMangaInsert(w http.ResponseWriter, r *http.Request) {

	var m kaoriManga.Manga

	mappa := r.Context().Value("values").(ContextValues)

	idAnilist := filepath.Base(r.URL.Path)

	//Read client data
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	num, err := strconv.Atoi(idAnilist)
	if err != nil {
		return
	}

	if m.Id != 0 {

		if m.Id != num {
			kaoriUtils.PrintErr(w, "Error, id of body request and id in the URL don't match")
			return
		}

	} else {
		m.Id = num
	}

	err = m.SendToDatabaseNR(kaoriMangaDB.Client.C, kaoriMangaDB.Client.Ctx)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	fmt.Println("MANGA:", m)

	kaoriUtils.PrintOk(w)

}
 */

//API ADMIN COMMAND

func ApiCommandRestart(w http.ResponseWriter, r *http.Request){

	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandRestart", "Error to send signal: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

}

func ApiCommandShutdown(w http.ResponseWriter, r *http.Request) {
	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandShutdown", "Error to send signal: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
}

func ApiCommandForcedShutdown(w http.ResponseWriter, r *http.Request) {
	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		kaoriLog.PrintLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandForcedShutdown", "Error to send signal: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
}
