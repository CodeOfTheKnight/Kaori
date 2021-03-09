package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	//"github.com/gorilla/mux"
)

var httpAddr = flag.String("http", ":8090", "Listen address")

func main() {
	flag.Parse()
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", ServeHTTPS)
	log.Fatal(http.ListenAndServeTLS(*httpAddr, "cert.pem", "key.pem", nil))
}

func ServeHTTPS(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if pusher, ok := w. (http.Pusher); ok {

		if err := pusher.Push("/static/app.js", nil); err != nil {
			log.Printf("Failed to push: %v", err)
		}
		if err := pusher.Push("/static/style.css", nil); err != nil {
			log.Printf("Failed to push: %v", err)
		}

	}
	fmt.Fprintf(w, "prova")
}
