package main

import (
	logger "github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

//Settings
const pathGui string = "./KaoriGui"
const nameDirGui string = "KaoriGui"
const certFile string = "cert/cert.pem"
const keyPem string = "cert/key.pem"
const port string = ":8020"

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
}

func main() {

	fs := http.FileServer(http.Dir(pathGui))

	mux := &http.ServeMux{}
	mux.HandleFunc("/", serveIndex)
	mux.Handle(endpointGui.String(), http.StripPrefix(string(filepath.Separator) + nameDirGui + string(filepath.Separator), fs))

	var handler http.Handler = mux
	handler = logRequestHandler(handler)

	server := http.Server{
		Addr:           port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(server.ListenAndServeTLS(os.Getenv("CERTIFICATE"), os.Getenv("KEY")))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {

	enableCors(w) //CORS POLICY

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

