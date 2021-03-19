package main

import (
	"github.com/gorilla/mux"
	logger "github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//Settings
const pathGui string = "./KaoriGui"
const pathTests string = "./tests"
const nameDirGui string = "KaoriGui"
const certFile string = "cert/cert.pem"
const keyPem string = "cert/key.pem"
const port string = ":8010"

func init() {

	file, err := os.OpenFile(logServer, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Log as JSON instead of the default ASCII formatter.
	logger.SetFormatter(&logger.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logger.SetOutput(file)

	//Set global variables
	os.Setenv("CERTIFICATE", filepath.FromSlash(certFile))
	os.Setenv("KEY", filepath.FromSlash(keyPem))
	os.Setenv("ACCESS_SECRET", "secret")
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
	routerAdd := router.PathPrefix(endApiAddData.String()).Subrouter()
	routerTest := router.PathPrefix(endpointTest.String()).Subrouter()

	//Rotte base
	router.HandleFunc(endpointRoot.String(), serveIndex)
	router.PathPrefix("/KaoriGui/").Handler(http.StripPrefix("/KaoriGui/", http.FileServer(http.Dir(pathGui))))

	//Rotte API AddData
	routerAdd.Path(endApiAddDataMusic.String()).HandlerFunc(ApiAddMusic).Methods(http.MethodPost)

	//Rotte API test
	routerTest.Use(authmiddleware)
	routerTest.PathPrefix(testFiles.String()).Handler(http.StripPrefix(path.Join(endpointTest.String(), testFiles.String()), http.FileServer(http.Dir(pathTests)))).Methods(http.MethodGet, http.MethodOptions)

	server := http.Server{
		Addr:           "0.0.0.0" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

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

		metadata, err := ExtractTokenMetadata(r)
		if err != nil {
			printInternalErr(w)
			return
		}

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