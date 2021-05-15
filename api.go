package main

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/firestore"
	"encoding/json"
	"fmt"
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

type ParamsInfo struct {
	Key      string
	Required bool
}

//API USER

func ApiUserInfo(w http.ResponseWriter, r *http.Request){

	var u User

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	//Get document
	document, err := kaoriUser.Client.c.Collection("User").
										Doc(mappa.Get("email")).
										Get(kaoriUser.Client.ctx)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error database: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	err = document.DataTo(&u)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error to create user struct: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	data, err := json.Marshal(u)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiUserInfo", "Error create JSON: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	w.Write(data)
}

//API USER SETTINGS

func ApiSettingsGet(w http.ResponseWriter, r *http.Request) {

	var s Settings

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "[Start] Get settings", 0)

	//Get document
	document, err := kaoriUser.Client.c.Collection("User").
		Doc(mappa.Get("email")).
		Get(kaoriUser.Client.ctx)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "Error database: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	userData := document.Data()

	err = mapstructure.Decode(userData["Settings"], &s)
	if err != nil {
		log.Println(err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "Error create JSON: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsGet", "[Done] Get settings", 0)

	w.Write(data)

}

func ApiSettingsSet(w http.ResponseWriter, r *http.Request){

	//var s Settings
	var u User
	var m map[string]interface{}

	mappa := r.Context().Value("values").(ContextValues)
	fmt.Println(mappa)

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "[Start] Change settings", 0)

	//Read precedent settings
	document, err := kaoriUser.Client.c.Collection("User").
										Doc(mappa.Get("email")).
										Get(kaoriUser.Client.ctx)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error database: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Parse document firebase in users struct
	err = document.DataTo(&u)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to conversion in settings: " + err.Error(), 1)
		printInternalErr(w)
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
		printInternalErr(w)
		return
	}

	fmt.Println("Settings:", u.Settings)

	//Check settings value
	if err = u.Settings.IsValid(); err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Data is invalid: " + err.Error(), 1)
		printErr(w, err.Error())
		return
	}

	//Send map to the database
	_, err =  kaoriUser.Client.c.Collection("User").
		Doc(mappa.Get("email")).
		Set(kaoriUser.Client.ctx, m, firestore.MergeAll)

	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error database: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Make JSON response
	data, err := json.Marshal(u.Settings)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "Error to create JSON: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

	printLog(mappa.Get("email"), mappa.Get("ip"), "ApiSettingsSet", "[Done] Change settings", 0)

	w.Write(data)
}

//API AUTH

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

	var u User

	// Declare a new MusicData struct.
	user := struct {
		Email          string `json:"email"`
		Username       string `json:"username"`
		Password       string `json:"password"`
		ProfilePicture string `json:"profilePicture,omitempty"`
	}{}

	ip := GetIP(r)

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		printErr(w, err.Error())
		return
	}

	u.Email = user.Email
	u.Username = user.Username
	u.Password = user.Password
	u.ProfilePicture = user.ProfilePicture

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
	signupField := cfg.Template.Mail["registration"]
	err = sendEmail(u.Email, signupField.Object, c, signupField.File)
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

	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error database: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	//Redirect to login
	data, err := parseTemplateHtml(cfg.Template.Html["redirect"], "https://"+cfg.Server.Host+cfg.Server.Port+endpointLogin.String())
	if err != nil {
		printLog(p["email"].(string), ip, "ApiConfirmSignup", "Error to create template: "+err.Error(), 1)
		printInternalErr(w)
		return
	}

	w.Write([]byte(data))
}

//API SERVICE

func ApiServiceAnime(w http.ResponseWriter, r *http.Request) {

	var a kaoriData.Anime
	//var eps []*Episode

	//GetIP
	//ip := GetIP(r)

	//Get anime id
	u := r.URL.Path
	id := filepath.Base(u)

	fmt.Println("ID:", id)

	a.Id = id

	err := a.GetAnimeFromDb(kaoriDataDB.Client.c, kaoriDataDB.Client.ctx)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(a)
}

//API ADD-DATA

func ApiAddMusic(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	//Extract token JWT
	tokenString := ExtractToken(r)
	metadata, err := ExtractAccessTokenMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		printLog("General", ip, "ApiConfigGet", "Error to extract access token metadata: " + err.Error(), 1)
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

//API ADMIN

func ApiConfigGet(w http.ResponseWriter, r *http.Request){

	ip := GetIP(r)

	//Extract token JWT
	tokenString := ExtractToken(r)
	metadata, err := ExtractAccessTokenMetadata(tokenString, cfg.Password.AccessToken)
	if err != nil {
		printLog("General", ip, "ApiConfigGet", "Error to extract access token metadata: " + err.Error(), 1)
		printErr(w, err.Error())
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

	var cfg2 Config
	cfg2 = *cfg

	mappa := r.Context().Value("values").(ContextValues)
	data, _ := io.ReadAll(r.Body)

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&cfg2)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Error to create JSON: " + err.Error(), 1)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Check configuration validity
	if err = cfg2.CheckConfig(); err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiConfigSet", "Config not valid: " + err.Error(), 1)
		printErr(w, "Config not valid: " + err.Error())
		return
	}

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
		set := []ParamsInfo{
			{Key: "func", Required: false},
			{Key: "ip", Required: false},
			{Key: "user", Required: false},
			{Key: "msg", Required: false},
			{Key: "level", Required: false},
			{Key: "time", Required: false},
			{Key: "order", Required: true},
		}

		params, err = getParams(set, r)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to get params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

		if err = checkFiltersLogGet(set, params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

	} else {

		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to get JSON params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

		if err = checkFiltersLogPost(params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

	}


	f, err := os.Open("log/server.log.json")
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to open log file: "+ err.Error(),1 )
		printInternalErr(w)
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
			logsString, err = filterLog(logsString, filter, params[filter].(string))
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
			printInternalErr(w)
			return
		}

		logs = append(logs, &sl)

	}

	data, err := json.Marshal(logs)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of slice log: "+ err.Error(),1 )
		printInternalErr(w)
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
		set := []ParamsInfo{
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

		params, err = getParams(set, r)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to get params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

		if err = checkFiltersLogGet(set, params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

	} else {
		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to get JSON params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}

		if err = checkFiltersLogPost(params); err != nil {
			printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error params: "+err.Error(), 1)
			printErr(w, err.Error())
			return
		}
	}


	f, err := os.Open("log/connection.log.json")
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogConnection", "Error to open log file: "+ err.Error(),1 )
		printInternalErr(w)
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
			logsString, err = filterLog(logsString, filter, params[filter].(string))
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
			printInternalErr(w)
			return
		}

		logs = append(logs, &hl)

	}

	data, err := json.Marshal(logs)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiLogServer", "Error to create JSON of slice log: "+ err.Error(),1 )
		printInternalErr(w)
		return
	}

	w.Write(data)

}

func ApiAnimeInsert(w http.ResponseWriter, r *http.Request) {

	var a kaoriData.Anime

	mappa := r.Context().Value("values").(ContextValues)

	//Read client data
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		printErr(w, err.Error())
		return
	}



	err = a.SendToDb(kaoriDataDB.Client.c, kaoriDataDB.Client.ctx)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiAnimeInsert", "Error to get params: "+err.Error(), 1)
		printErr(w, err.Error())
		return
	}

	fmt.Println("ARCHIVE:", a.Episodes[0].Videos)
}

//API ADMIN COMMAND

func ApiCommandRestart(w http.ResponseWriter, r *http.Request){

	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandRestart", "Error to send signal: " + err.Error(), 1)
		printInternalErr(w)
		return
	}

}

func ApiCommandShutdown(w http.ResponseWriter, r *http.Request) {
	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandShutdown", "Error to send signal: " + err.Error(), 1)
		printInternalErr(w)
		return
	}
}

func ApiCommandForcedShutdown(w http.ResponseWriter, r *http.Request) {
	mappa := r.Context().Value("values").(ContextValues)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		printLog(mappa.Get("email"), mappa.Get("ip"), "ApiCommandForcedShutdown", "Error to send signal: " + err.Error(), 1)
		printInternalErr(w)
		return
	}
}