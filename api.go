package main

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/firestore"
	"encoding/json"
	"fmt"
	kaoriUser "github.com/CodeOfTheKnight/Kaori/kaoriData/user"
	"github.com/CodeOfTheKnight/Kaori/kaoriJwt"
	"github.com/CodeOfTheKnight/Kaori/kaoriMail"
	"github.com/CodeOfTheKnight/Kaori/kaoriSettings"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"github.com/CodeOfTheKnight/kaoriData"
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
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	err = document.DataTo(&u)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error to create user struct: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	data, err := json.Marshal(u)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error create JSON: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)
}

//API USER SETTINGS

func ApiSettingsGet(w http.ResponseWriter, r *http.Request) {

	var s kaoriUser.Settings

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "[Start] Get settings", 0)

	//Get document
	document, err := kaoriUserDB.Client.C.Collection("User").
		Doc(mappa.Get("email")).
		Get(kaoriUserDB.Client.Ctx)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	fmt.Println("STRUCT:", document.Data())

	userData := document.Data()

	err = mapstructure.Decode(userData["Settings"], &s)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("SETTING:", s)

	data, err := json.Marshal(s)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "Error create JSON: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "[Done] Get settings", 0)

	w.Write(data)

}

func ApiSettingsSet(w http.ResponseWriter, r *http.Request){

	//var s Settings
	var u kaoriUser.User
	var m map[string]interface{}

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "[Start] Change settings", 0)

	//Read precedent settings
	document, err := kaoriUserDB.Client.C.Collection("User").
										Doc(mappa.Get("email")).
										Get(kaoriUserDB.Client.Ctx)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Parse document firebase in users struct
	err = document.DataTo(&u)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to conversion in settings: " + err.Error(), 1)
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
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to conversion in a map: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	fmt.Println("MAPPA:", m)

	fmt.Println("Settings:", u.Settings)

	//Check settings value
	if err = u.Settings.IsValid(); err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Data is invalid: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	//Send map to the database
	_, err =  kaoriUserDB.Client.C.Collection("User").
		Doc(mappa.Get("email")).
		Set(kaoriUserDB.Client.Ctx, map[string]interface{}{"Settings": m}, firestore.MergeAll)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error database: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Make JSON response
	data, err := json.Marshal(u.Settings)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to create JSON: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "[Done] Change settings", 0)

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
		printLog("General", ip, "ApiUserExist", "Error to get params: "+err.Error(), 1)
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
	tokens, err := kaoriJwt.GenerateTokenPair(params.Email)
	if err != nil {
		printLog("General", ip, "ApiLogin", "Error to generate token pair: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Set Cookies
	err = kaoriUtils.SetCookies(w, tokens["RefreshToken"].Token, cfg.Jwt.Iss, cfg.Password.Cookies)
	if err != nil {
		printLog("General", ip, "ApiLogin", "Error to set cookies: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Save refresh token in the database
	data := tokens["RefreshToken"].Fields
	exp := data["exp"].(time.Time).Unix()

	rf, ok := data["refreshId"].(string)
	if !ok {
		printLog("General", ip, "ApiLogin", "Error, the field with \"refreshId\" key doesn't exist", 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	_, err = kaoriUserDB.Client.C.Collection("User").Doc(params.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUserDB.Client.Ctx, map[string]int64{"exp": exp})

	if err != nil {
		printLog("General", ip, "ApiLogin", "Database operation error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Create JSON
	fields := tokens["AccessToken"].Fields
	exp, ok = fields["exp"].(int64)
	if !ok {
		printLog("General", ip, "ApiLogin", "Error, the field with \"AccessToken\" key doesn't exist", 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	data2, err := json.Marshal(map[string]string{
		"AccessToken": tokens["AccessToken"].Token,
		"Expiration":  fmt.Sprint(exp),
	})
	if err != nil {
		printLog("General", ip, "ApiLogin", "Error to create AccessToken JSON: "+err.Error(), 1)
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
		printLog("General", ip, "ApiRefresh", "Error to get cookies: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "")
		return
	}

	token := cookieData["RefreshToken"]

	//Extract data from token
	data, err := kaoriJwt.ExtractRefreshTokenMetadata(token, cfg.Password.RefreshToken)
	if err != nil {
		printLog("General", ip, "ApiRefresh", "Extract refresh token error: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, "Cookies Not valid")
		return
	}

	//Check validity
	if !kaoriJwt.VerifyRefreshToken(data.Email, data.RefreshId) {
		printLog(data.Email, ip, "ApiRefresh", "Token not valid", 2)
		kaoriUtils.PrintErr(w, "Token not valid")
		return
	}

	if !kaoriJwt.VerifyExpireDate(data.Exp) {
		printLog(data.Email, ip, "ApiRefresh", "Token expired", 2)
		http.Error(w, `{"code": 401, "msg": "Token expired! Login required!"}`, http.StatusUnauthorized)
		return
	}

	//Remove old refresh token
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(data.RefreshId).
		Delete(kaoriUserDB.Client.Ctx)

	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Database connection error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Check token expired
	err = kaoriJwt.CheckOldsToken(data.Email)
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Check old tokens error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Generate new token pair
	tokens, err := kaoriJwt.GenerateTokenPair(data.Email)
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Generate token pair error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Set Cookies
	err = kaoriUtils.SetCookies(w, tokens["RefreshToken"].Token, cfg.Jwt.Iss, cfg.Password.Cookies)
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Set cookie error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Save refresh token in the database
	data2 := tokens["RefreshToken"].Fields
	exp := data2["exp"].(time.Time).Unix()

	rf, ok := data2["refreshId"].(string)
	if !ok {
		printLog(data.Email, ip, "ApiRefresh", `The field with key "refreshId" doesn't exist'`, 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	_, err = kaoriUserDB.Client.C.Collection("User").Doc(data.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUserDB.Client.Ctx, map[string]int64{"exp": exp})

	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Database connection error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Create JSON
	fields := tokens["AccessToken"].Fields
	exp, ok = fields["exp"].(int64)
	if !ok {
		printLog(data.Email, ip, "ApiRefresh", `The field with key "exp" doesn't exist'`, 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Create JSON
	jsonData, err := json.Marshal(map[string]string{
		"AccessTocken": tokens["AccessToken"].Token,
		"Expiration":   fmt.Sprint(exp),
	})
	if err != nil {
		printLog(data.Email, ip, "ApiRefresh", "Create JSON error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(jsonData)
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
		printLog("General", ip, "ApiSignUp", "AddNewUser error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Generate tokens
	tokens, err := kaoriJwt.GenerateTokenPair(u.Email)
	if err != nil {
		printLog("General", ip, "ApiSignUp", "Error to generate token pair: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
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
		kaoriUtils.PrintInternalErr(w)
		return
	}

	_, err = kaoriUserDB.Client.C.Collection("User").Doc(u.Email).
		Collection("RefreshToken").Doc(rf).
		Set(kaoriUserDB.Client.Ctx, map[string]int64{"exp": exp})

	if err != nil {
		printLog("General", ip, "ApiSignUp", "Database error: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
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
		printLog("General", ip, "ApiSignUp", "Error to send mail: "+err.Error(), 1)
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
		printLog("General", ip, "ApiConfirmSignup", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	if !kaoriJwt.VerifyRefreshToken(p["email"].(string), p["id"].(string)) {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Warning to verify refresh token: Token not valid", 2)
		kaoriUtils.PrintErr(w, "Token not valid!")
		return
	}

	//Remove old refresh token
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(p["email"].(string)).
		Collection("RefreshToken").Doc(p["id"].(string)).
		Delete(kaoriUserDB.Client.Ctx)

	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Set user to active
	_, err = kaoriUserDB.Client.C.Collection("User").Doc(p["email"].(string)).
		Set(kaoriUserDB.Client.Ctx, map[string]bool{"IsActive": true}, firestore.MergeAll)

	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Redirect to login
	data, err := kaoriUtils.ParseTemplateHtml(cfg.Template.Html["redirect"], "https://"+cfg.Server.Host+cfg.Server.Port+endpointLogin.String())
	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error to create template: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write([]byte(data))
}

//API SERVICE

func ApiServiceAnime(w http.ResponseWriter, r *http.Request) {

	var a kaoriData.Anime
	var err error

	//GetIP
	ip := kaoriUtils.GetIP(r)

	//Get anime id
	id := filepath.Base(r.URL.Path)
	a.Id, err = strconv.Atoi(id)
	if err != nil {
		printLog("General", ip, "ServiceAnime", err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	err = a.GetAnimeFromDb(kaoriDataDB.Client.C, kaoriDataDB.Client.Ctx)
	if err != nil {
		printLog("General", ip, "ServiceAnime", err.Error(), 1)
		http.Error(w, "Anime not found", http.StatusNotFound)
		return
	}

	data, _ := json.Marshal(a)

	//TODO: Generate template
	w.Write(data)
}

func ApiServiceManga(w http.ResponseWriter, r *http.Request) {

}

//API ADD-DATA

func ApiAddMusic(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	//Extract token JWT
	tokenString := kaoriJwt.ExtractToken(r)
	metadata, err := kaoriJwt.ExtractAccessTokenMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		printLog("General", ip, "ApiConfigGet", "Error to extract access token metadata: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	// Declare a new MusicData struct.
	var md kaoriData.MusicData

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
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	//Check if it already exist
	_, err = kaoriTmpDB.Client.C.Collection(md.Type).
		Doc(strconv.Itoa(md.IdAnilist)).
		Get(kaoriTmpDB.Client.Ctx)
	if err == nil {
		printLog(metadata.Email, ip, "ApiAddMusic", fmt.Sprintf("Warning item with id=%s already exist: ", md.IdAnilist), 2)
		http.Error(w, `{"code": 409, "msg": "The track already exists"}`, http.StatusConflict)
		return
	}

	md.GetNameAnime()
	err = md.NormalizeName(cfg.Template.Music["name"])
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Error to normalize name: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Upload
	err = md.UploadTemporaryFile()
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Error to upload temporary file: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	//Add to database
	err = md.AddDataToTmpDatabase(kaoriTmpDB)
	if err != nil {
		printLog(metadata.Email, ip, "ApiAddMusic", "Error to add music data to the temp database: "+err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
}

//API ADMIN

func ApiConfigGet(w http.ResponseWriter, r *http.Request){

	ip := kaoriUtils.GetIP(r)

	//Extract token JWT
	tokenString := kaoriJwt.ExtractToken(r)
	metadata, err := kaoriJwt.ExtractAccessTokenMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		printLog("General", ip, "ApiConfigGet", "Error to extract access token metadata: " + err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		printLog(metadata.Email, ip, "ApiConfigGet", "Error to create JSON of config struct: " + err.Error(), 1)
	}

	w.Write(data)

	printLog(metadata.Email, ip, "ApiConfigGet", "The user has read the settings", 0)
}

func ApiConfigSet(w http.ResponseWriter, r *http.Request){

	var cfg2 kaoriSettings.Config
	cfg2 = *cfg

	mappa := r.Context().Value("values").(ContextValues)
	data, _ := io.ReadAll(r.Body)

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&cfg2)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Error to create JSON: " + err.Error(), 1)
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
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Error to write config: "+err.Error(), 1)
		http.Error(w, fmt.Sprintf("{\"code\": 500, \"msg\": \"%s\"}\n", strings.Join(wdone, "\n")), http.StatusInternalServerError)
		return
	}

	//TODO: Change file config
	//TODO: Reboot
	fmt.Println(cfg2)

	w.Write([]byte(fmt.Sprintf("{\"code\": 200, \"msg\": \"\n%s\"}\n", strings.Join(wdone, "\n"))))
}

func ApiLogServer(w http.ResponseWriter, r *http.Request){

	var logs []*ServerLog
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
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to get params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogGet(set, params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

	} else {

		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to get JSON params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogPost(params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

	}


	f, err := os.Open("log/server.log.json")
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to open log file: "+ err.Error(),1 )
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

		var sl ServerLog

		err = json.Unmarshal([]byte(item), &sl)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of single log: "+ err.Error(),1 )
			kaoriUtils.PrintInternalErr(w)
			return
		}

		logs = append(logs, &sl)

	}

	data, err := json.Marshal(logs)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of slice log: "+ err.Error(),1 )
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)
}

func ApiLogConnection(w http.ResponseWriter, r *http.Request){

	var logsString []string
	var logs []*HTTPReqInfo
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
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to get params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogGet(set, params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

	} else {
		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to get JSON params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}

		if err = kaoriUtils.CheckFiltersLogPost(params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error params: "+err.Error(), 1)
			kaoriUtils.PrintErr(w, err.Error())
			return
		}
	}


	f, err := os.Open("log/connection.log.json")
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to open log file: "+ err.Error(),1 )
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

		var hl HTTPReqInfo

		err = json.Unmarshal([]byte(item), &hl)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to create JSON of single log: "+ err.Error(),1 )
			kaoriUtils.PrintInternalErr(w)
			return
		}

		logs = append(logs, &hl)

	}

	data, err := json.Marshal(logs)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of slice log: "+ err.Error(),1 )
		kaoriUtils.PrintInternalErr(w)
		return
	}

	w.Write(data)

}

func ApiAnimeInsert(w http.ResponseWriter, r *http.Request) {

	var a kaoriData.Anime

	mappa := r.Context().Value("values").(ContextValues)

	idAnilist := filepath.Base(r.URL.Path)

	//Read client data
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	num, err := strconv.Atoi(idAnilist)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

	if a.Id != 0 {

		if a.Id != num {
			kaoriUtils.PrintErr(w, "Error, id of body request and id in the URL don't match")
			return
		}

	} else {
		a.Id = num
	}

	err = a.SendToDb(kaoriDataDB.Client.C, kaoriDataDB.Client.Ctx)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	kaoriUtils.PrintOk(w)
}

func ApiMangaInsert(w http.ResponseWriter, r *http.Request) {

	var m kaoriData.Manga

	mappa := r.Context().Value("values").(ContextValues)

	idAnilist := filepath.Base(r.URL.Path)

	//Read client data
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	if m.Id != "" {
		if m.Id != idAnilist {
			kaoriUtils.PrintErr(w, "Error, id of body request and id in the URL don't match")
			return
		}
	} else {
		m.Id = idAnilist
	}

	err = m.SendToDatabase(kaoriMangaDB.Client.C, kaoriMangaDB.Client.Ctx)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		kaoriUtils.PrintErr(w, err.Error())
		return
	}

	fmt.Println("MANGA:", m)

	kaoriUtils.PrintOk(w)

}

//API ADMIN COMMAND

func ApiCommandRestart(w http.ResponseWriter, r *http.Request){

	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandRestart", "Error to send signal: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}

}

func ApiCommandShutdown(w http.ResponseWriter, r *http.Request) {
	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandShutdown", "Error to send signal: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
}

func ApiCommandForcedShutdown(w http.ResponseWriter, r *http.Request) {
	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandForcedShutdown", "Error to send signal: " + err.Error(), 1)
		kaoriUtils.PrintInternalErr(w)
		return
	}
}