package main

import (
	"fmt"
	"github.com/gorilla/mux"
	logger "github.com/sirupsen/logrus"
//	"google.golang.org/api/gmail/v1"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//Settings
const baseUrl string = "ec2-13-58-107-13.us-east-2.compute.amazonaws.com"
const pathGui string = "./KaoriGui"
const pathTests string = "./tests"
const nameDirGui string = "KaoriGui"
const certFile string = "cert/cert.pem"
const keyPem string = "cert/key.pem"
const port string = ":8012"
const adminToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJLYW9yaVN0cmVhbSIsImlhdCI6MTYxNjUxMDIyMSwiZXhwIjoxNjE3MTA3Nzg4LCJjb21wYW55IjoiQ29kZU9mVGhlS25pZ2h0IiwiZW1haWwiOiJ3YXRhc2hpd2F5dXJpZGFpc3VraUBnbWFpbC5jb20iLCJwZXJtaXNzaW9uIjoidWN0YSJ9.lf4KbMv2TK-eFeiGHS_jlIf5OMFLs18EvTbAWKt-Ef4"

func init() {

	//SET LOGGER
	file, err := os.OpenFile(logServer, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Log as JSON instead of the default ASCII formatter.
	logger.SetFormatter(&logger.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logger.SetOutput(file)

	//SET DATABASES
	kaoriTmp, err = NewDatabase("kaori-504c3", "database/kaori-504c3-firebase-adminsdk-5apba-f66a21203e.json")
	if err != nil {
		panic(err)
	}

	kaoriUser, err = NewDatabase("kaoriuser-fbae1", "database/kaoriuser-fbae1-firebase-adminsdk-1dfp8-d23953ad42.json")
	if err != nil {
		panic(err)
	}

	//SET GLOBAL VARIABLES
	os.Setenv("CERTIFICATE", filepath.FromSlash(certFile))
	os.Setenv("KEY", filepath.FromSlash(keyPem))
	os.Setenv("ACCESS_SECRET", "secret")
	os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf")
}

/*PRIVILEGI: ucta
	u: User = Accesso base
	c: Creator = Accesso alle api di verifica dati aggiunti da utenti
	t: Tester = Accesso alle api di test
	a: Admin = tutti gli accessi
*/

func main() {

	//Creazione router
	router := mux.NewRouter()
	router.Use(enableCors) //CORS middleware
	routerAdd := router.PathPrefix(endpointAddData.String()).Subrouter()
	routerTest := router.PathPrefix(endpointTest.String()).Subrouter()
	routerUser := router.PathPrefix(endpointUser.String()).Subrouter()
	routerAuth := router.PathPrefix(endpointAuth.String()).Subrouter()

	//Rotte base
	router.HandleFunc(endpointRoot.String(), serveIndex)
	router.PathPrefix("/KaoriGui/").Handler(http.StripPrefix("/KaoriGui/", http.FileServer(http.Dir(pathGui))))

	//Rotte API AddData
	routerAdd.Path(addDataMusic.String()).HandlerFunc(ApiAddMusic).Methods(http.MethodPost)

	//Rotte API User
	routerUser.Path(userExist.String()).HandlerFunc(ApiUserExist).Methods(http.MethodGet)

	//Rotte API Auth
	routerAuth.Path(authLogin.String()).HandlerFunc(ApiLogin).Methods(http.MethodPost)
	routerAuth.Path(authRefresh.String()).HandlerFunc(ApiRefresh).Methods(http.MethodGet)
	routerAuth.Path(authSignUp.String()).HandlerFunc(ApiSignUp).Methods(http.MethodPost)
	routerAuth.Path(authConfirmSignUp.String()).HandlerFunc(ApiConfirmSignUp).Methods(http.MethodGet)

	//Rotte API test
	routerTest.Use(authmiddleware)
	routerTest.PathPrefix(testFiles.String()).Handler(http.StripPrefix(path.Join(endpointTest.String(), testFiles.String()), http.FileServer(http.Dir(pathTests)))).Methods(http.MethodGet, http.MethodOptions)

	server := http.Server{
		Addr:           "0.0.0.0" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go mailDetector()

	log.Fatal(server.ListenAndServeTLS(os.Getenv("CERTIFICATE"), os.Getenv("KEY")))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		ip := GetIP(r)

		//Server push
		files, err := ls(filepath.ToSlash(nameDirGui + "/"))

		if pusher, ok := w.(http.Pusher); ok {
			// Push is supported.
			for _, file := range files {
				if err := pusher.Push(filepath.ToSlash(path.Join("/", file)), nil); err != nil {
					log.Printf("Failed to push: %v", err)
				}
			}
		}

		content, err := os.ReadFile(filepath.ToSlash(filepath.Join(nameDirGui, "/home.html")))
		if err != nil {
			log.Println(ip, err)
			printInternalErr(w)
			return
		}

		w.Write(content)

	} else {
		printErr(w, "")
		return
	}
}

func authmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := ExtractToken(r)
		metadata, err := ExtractAccessTokenMetadata(tokenString, os.Getenv("ACCESS_SECRET"))
		if err != nil {
			printErr(w, err.Error())
			return
		}

		fmt.Println(metadata.Permission)

		if !strings.Contains(metadata.Permission, "a"){
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("HTTP 403- Forbidden"))
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)

	})
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

//POTREBBE DARE ERRORE NEL CASO HA RAGGIUNTO IL MASSIMO DI MAIL LETTE/SCRITTE
//NEL CASO COMMENTARE POI SISTEMERÒ
func mailDetector() {

/*
	var srv Service

	config, err := readMailConfig(gmail.MailGoogleComScope)
	if err != nil {
		log.Fatal(err)
	}
	client := getClient(config)

	srv, err = NewService(client)

	for x := range time.Tick(10 * time.Second) {
		srv.WaitMail(x)
	}

*/
}

func (s *Service) WaitMail(t time.Time) {

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
}
