package main

import (
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/mux"
	logger "github.com/sirupsen/logrus"
	//	"google.golang.org/api/gmail/v1"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var cfg *Config

//Settings
const indexFile string = "home.html"
const loginFile string = "login.html"
var lmt *limiter.Limiter

func init() {
	var err error

	//READ CONFIG
	cfg, err = NewConfig()
	if err != nil {
		log.Fatalln(err)
		return
	}

	//VALIDATE CONFIG
	err = cfg.CheckConfig()
	if err != nil {
		printLog("Server", "", "init", "Error to validate config: " + err.Error(), 1)
		return
	}

	//SET LOGGER
	file, err := os.OpenFile(cfg.Logger.Server, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Log as JSON instead of the default ASCII formatter.
	logger.SetFormatter(&logger.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logger.SetOutput(file)

	printLog("Server", "", "init", "Setting Database start", 0)

	//SET DATABASES
	kaoriTmp, err = NewDatabase(cfg.Database[0].ProjectId, cfg.Database[0].Key)
	if err != nil {
		printLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	kaoriUser, err = NewDatabase(cfg.Database[1].ProjectId, cfg.Database[1].Key)
	if err != nil {
		printLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	printLog("Server", "", "init", "Setting Database done", 0)

	//Create limiter middleware
	lmt = tollbooth.NewLimiter(float64(cfg.Server.Limiter), nil)
	lmt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).SetMethods([]string{"POST", "GET"})
	lmt.SetMessage("{\"code\": 429, \"msg\": \"Too many request!\"}\n")

	printLog("Server", "", "init", "Setting Limiter done", 0)
}

func main() {

	//Setting auth middleware
	userAuthMiddleware := NewAuthMiddlewarePerm(UserPerm)
	//creatorAuthMiddleware := NewAuthMiddlewarePerm(CreatorPerm)
	testerAuthMiddleware := NewAuthMiddlewarePerm(TesterPerm)
	adminAuthMiddleware := NewAuthMiddlewarePerm(AdminPerm)

	//Creazione router
	router := mux.NewRouter()
	router.Use(enableCors) //CORS middleware

	//Creazione subrouter per api di aggiunta dati
	routerAdd := router.PathPrefix(endpointAddData.String()).Subrouter()
	routerAdd.Use(userAuthMiddleware.authmiddleware)

	//Creazione subrouter per api di test
	routerTest := router.PathPrefix(endpointTest.String()).Subrouter()
	routerTest.Use(testerAuthMiddleware.authmiddleware)

	//Creazione subrouter per api di utente
	routerUser := router.PathPrefix(endpointUser.String()).Subrouter()
	routerUser.Use(userAuthMiddleware.authmiddleware)

		//Creazione subrouter per api di settings
		routerSettings := routerUser.PathPrefix(userSettings.String()).Subrouter()
		routerSettings.Use(userAuthMiddleware.authmiddleware)

	//Creazione subrouter per api di autenticazione
	routerAuth := router.PathPrefix(endpointAuth.String()).Subrouter()

	//Creazione subrouter per api di amministratore
	routerAdmin := router.PathPrefix(endpointAdmin.String()).Subrouter()
	routerAdmin.Use(adminAuthMiddleware.authmiddleware)

	//Rotte base
	router.HandleFunc(endpointRoot.String(), serveIndex)
	router.HandleFunc(endpointLogin.String(), serveLogin)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.Server.Gui)))

	//Rotte API AddData
	routerAdd.Path(addDataMusic.String()).HandlerFunc(ApiAddMusic).Methods(http.MethodPost)

	//Rotte API Auth
	routerAuth.Path(authLogin.String()).HandlerFunc(ApiLogin).Methods(http.MethodPost)
	routerAuth.Path(authRefresh.String()).HandlerFunc(ApiRefresh).Methods(http.MethodGet)
	routerAuth.Path(authSignUp.String()).HandlerFunc(ApiSignUp).Methods(http.MethodPost)
	routerAuth.Path(authConfirmSignUp.String()).HandlerFunc(ApiConfirmSignUp).Methods(http.MethodGet)
	routerAuth.Path(authUserExist.String()).HandlerFunc(ApiUserExist).Methods(http.MethodGet)

	//Rotte API test
	routerTest.PathPrefix(testFiles.String()).Handler(
		http.StripPrefix(
			path.Join(
				endpointTest.String(),
				testFiles.String(),
			),
			http.FileServer(http.Dir(cfg.Server.Test)),
		),
	).Methods(http.MethodGet, http.MethodOptions)

	//Rotte API user
	routerUser.Path(userInfo.String()).HandlerFunc(ApiUserInfo).Methods(http.MethodGet)

		//Rotte API settings
		routerSettings.Path(settingsGet.String()).HandlerFunc(ApiSettingsGet).Methods(http.MethodGet)
		routerSettings.Path(settingsSet.String()).HandlerFunc(ApiSettingsSet).Methods(http.MethodPost)

	//Rotte API admin
	routerAdmin.Path(adminConfigGet.String()).HandlerFunc(ApiConfigGet).Methods(http.MethodGet)
	routerAdmin.Path(adminConfigSet.String()).HandlerFunc(ApiConfigSet).Methods(http.MethodPost)

	fmt.Println("https://" + cfg.Server.Host + cfg.Server.Port + endpointAdmin.String() + adminConfigGet.String())

	//Add logger middleware
	var handler http.Handler = router
	handler = tollbooth.LimitHandler(lmt, handler)
	handler = logRequestHandler(handler)

	server := http.Server{
		Addr:           "0.0.0.0" + cfg.Server.Port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go mailDetector()

	printLog("Server", "", "main", "Start Server", 0)

	err := server.ListenAndServeTLS(cfg.Server.Ssl.Certificate, cfg.Server.Ssl.Key)
	if err != nil {
		printLog("Server", "", "main", "Server crash", 1)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		ip := GetIP(r)

		//Server push
		files, err := ls(filepath.ToSlash(cfg.Server.Gui))

		if pusher, ok := w.(http.Pusher); ok {
			// Push is supported.
			for _, file := range files {
				if err := pusher.Push(filepath.ToSlash(path.Join("/", file)), nil); err != nil {
					printLog("General", ip, "ServeHttp", fmt.Sprint("Failed to push: ", err), 1)
					return
				}
			}
		}

		content, err := os.ReadFile(filepath.ToSlash(filepath.Join(cfg.Server.Gui, indexFile)))
		if err != nil {
			printLog("General", ip, "ServeHttp", "Error to open index file", 1)
			printInternalErr(w)
			return
		}

		w.Write(content)

	} else {
		printErr(w, "")
		return
	}
}

func serveLogin(w http.ResponseWriter, r *http.Request) {

	ip := GetIP(r)

	data, err := os.ReadFile(filepath.Join(cfg.Server.Gui, loginFile))
	if err != nil {
		printLog("General", ip, "serveLogin", "Error to open file: "+loginFile, 1)
		printInternalErr(w)
		return
	}

	w.Write(data)
}

//POTREBBE DARE ERRORE NEL CASO HA RAGGIUNTO IL MASSIMO DI MAIL LETTE/SCRITTE
//NEL CASO COMMENTARE POI SISTEMERÒ
func mailDetector() {

	/*
		var srv Service

		conf, err := readMailConfig(gmail.MailGoogleComScope)
		if err != nil {
			log.Fatal(err)
		}
		client := getClient(conf)

		srv, err = NewService(client)

		for x := range time.Tick(10 * time.Second) {
			srv.WaitMail(x)
		}

	*/
}

/*func (s *Service) WaitMail(t time.Time) {

	err := s.GetMails()
	if err != nil {
		log.Fatal(err)
	}
	s.UnreadMails() //Prende solo mail non lette

	//Aggiunge tutte le mail di tipo "ADD_DATA" leggendo l'oggetto delle mail
	err = s.AddDataMails()
	if err != nil {
		log.Fatal(err)
	}


	for idm, mail := range s.Mails{

		if mail.IsUser() {
			music, err := mail.ParseMailMusic()
			if err != nil {
				err = sendEmail(mail.Head[4].Value, "Dati Errati", nil, filepath.Join(emailTemplate, "datiErrati.txt"))
				if err != nil {
					log.Println(err)
				}
				log.Println(err)
				continue
			}

			err = music.CheckError()
			if err != nil {
				log.Println(err)
				err = sendEmail(mail.Head[4].Value, "Dati Errati", nil, filepath.Join(emailTemplate, "datiErrati.txt"))
				if err != nil {
					log.Println(err)
				}
				log.Println(err)
				continue
			}

			music.GetNameAnime()

			err = music.NormalizeName()
			if err != nil {
				err = sendEmail(mail.Head[4].Value, "Problema Del Server", nil, filepath.Join(emailTemplate, "serverProblem.txt"))
				if err != nil {
					log.Println(err)
				}
				//TODO: LOGGER
				log.Println(err)
				continue
			}

			//Upload
			err = music.UploadTemporaryFile()
			if err != nil {
				err = sendEmail(mail.Head[4].Value, "Problema Del Server", nil, filepath.Join(emailTemplate, "serverProblem.txt"))
				if err != nil {
					log.Println(err)
				}
				log.Println(err)
				continue
			}

			//TODO: Verificare che non sia già nel database

			//Add to database
			err = music.AddDataToTmpDatabase()
			if err != nil {
				err = sendEmail(mail.Head[4].Value, "Problema Del Server", nil, filepath.Join(emailTemplate, "serverProblem.txt"))
				if err != nil {
					log.Println(err)
				}
				log.Println(err)
				continue
			}

			err = sendEmail(mail.Head[4].Value, "Aggiunta Con Successo", music, filepath.Join(emailTemplate, "aggiuntaCorretta.txt"))
			if err != nil {
				log.Println(err)
				continue
			}

			//ELIMINAZIONE MAIL
			delreq := s.Users.Messages.Delete("me", idm)
			err = delreq.Do()
			if err != nil {
				log.Fatal(err)
			}
		}

		//Se non è utente
		err = sendEmail(mail.Head[4].Value, "Servizio Negato", nil, filepath.Join(emailTemplate, "nonUtente.txt"))
		if err != nil {
			log.Println(err)
		}
		log.Println(err)
	}
}*/
