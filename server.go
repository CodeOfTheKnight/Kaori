package main

import (
	"github.com/felixge/httpsnoop"
	logger "github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"time"
)

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

}

func main() {

	fs := http.FileServer(http.Dir("./KaoriGui"))

	mux := &http.ServeMux{}
	mux.HandleFunc("/", serveIndex)
	mux.Handle("/KaoriGui/", http.StripPrefix("/KaoriGui/", fs))

	var handler http.Handler = mux
	handler = logRequestHandler(handler)

	server := http.Server{
		Addr:           ":8090",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(server.ListenAndServeTLS("cert/cert.pem", "cert/key.pem"))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {

	enableCors(w) //CORS POLICY

	if r.Method == http.MethodGet {

		ip := GetIP(r)

		content, err := os.ReadFile("KaoriGui/home.html")
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

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ri := &HTTPReqInfo{
			Method:    r.Method,
			Url:       r.URL.String(),
			Referer:   r.Header.Get("Referer"),
			UserAgent: r.Header.Get("User-Agent"),
			Data:      time.Now().Unix(),
		}

		ri.Ipaddr = GetIP(r)

		// this runs handler h and captures information about
		// HTTP request
		m := httpsnoop.CaptureMetrics(h, w, r)

		ri.Code = m.Code
		ri.Size = m.Written
		ri.Duration = m.Duration
		ri.logHTTPReq()
	}
	return http.HandlerFunc(fn)
}

