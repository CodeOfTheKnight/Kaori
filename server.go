package main

import (
	"context"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/CodeOfTheKnight/Kaori/kaoriLog"
	"github.com/CodeOfTheKnight/Kaori/kaoriSettings"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	logger "github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	//	"google.golang.org/api/gmail/v1"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	kaoriTmpDB *kaoriDatabase.NoSqlDb
	kaoriUserDB *kaoriDatabase.NoSqlDb
	kaoriDataDB *kaoriDatabase.NoSqlDb
	//kaoriMangaDB *kaoriDatabase.NoSqlDb
	kaoriMangaDB *kaoriDatabase.SqlDb
	kaoriAnimeDB *kaoriDatabase.SqlDb
	cfg *kaoriSettings.Config
)

//Settings
const indexFile string = "home.html"
const loginFile string = "login.html"
const configPath string = "kaoriSrc/config/"
var lmt *limiter.Limiter

type Server struct {
	server   http.Server
	listener *net.TCPListener
	wg       sync.WaitGroup
	sig chan os.Signal
}

func NewServer(handler http.Handler) *Server {
	var s Server
	s.sig = make(chan os.Signal, 1)
	s.server = http.Server{
		Addr:           "0.0.0.0" + cfg.Server.Port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return &s
}

func init() {

	var err error

	//Check and change precedent Settings
	err = kaoriSettings.CheckPrecedentConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	//READ CONFIG
	cfg, err = kaoriSettings.NewConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	//VALIDATE CONFIG
	err = cfg.CheckConfig()
	if err != nil {
		log.Fatalln(err)
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

	kaoriLog.PrintLog("Server", "", "init", "Setting Database start", 0)

	//SET DATABASES NO Relational
	kaoriTmpDB, err = kaoriDatabase.NewNoSqlDb(cfg.Database.NonRelational[0].ProjectId, cfg.Database.NonRelational[0].Key)
	if err != nil {
		kaoriLog.PrintLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	kaoriUserDB, err = kaoriDatabase.NewNoSqlDb(cfg.Database.NonRelational[1].ProjectId, cfg.Database.NonRelational[1].Key)
	if err != nil {
		kaoriLog.PrintLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	kaoriDataDB, err = kaoriDatabase.NewNoSqlDb(cfg.Database.NonRelational[2].ProjectId, cfg.Database.NonRelational[2].Key)
	if err != nil {
		kaoriLog.PrintLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	/*
	kaoriMangaDB, err = kaoriDatabase.NewNoSqlDb(cfg.Database.NonRelational[3].ProjectId, cfg.Database.NonRelational[3].Key)
	if err != nil {
		kaoriLog.PrintLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}
	*/

	//SET DATABASE RELAZIONALE

	kaoriAnimeDB, err = kaoriDatabase.NewSqlDb(cfg.Database.Relational[0].Username, cfg.Database.Relational[0].Password, cfg.Database.Relational[0].Host, cfg.Database.Relational[0].Port, cfg.Database.Relational[0].Db, "mysql")
	if err != nil {
		kaoriLog.PrintLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	kaoriMangaDB, err = kaoriDatabase.NewSqlDb(cfg.Database.Relational[1].Username, cfg.Database.Relational[1].Password, cfg.Database.Relational[1].Host, cfg.Database.Relational[1].Port, cfg.Database.Relational[1].Db, "mysql")
	if err != nil {
		kaoriLog.PrintLog("Server", "", "init", err.Error(), 1)
		panic(err)
	}

	kaoriLog.PrintLog("Server", "", "init", "Setting Database done", 0)

	//Create limiter middleware
	lmt = tollbooth.NewLimiter(float64(cfg.Server.Limiter), nil)
	lmt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).SetMethods([]string{"POST", "GET"})
	lmt.SetMessage("{\"code\": 429, \"msg\": \"Too many request!\"}\n")

	kaoriLog.PrintLog("Server", "", "init", "Setting Limiter done", 0)
}

func main() {

	pid := os.Getpid()

	//Create router with endpoints setted
	router := RouterInit()

	//Add logger middleware
	var handler http.Handler = router
	handler = tollbooth.LimitHandler(lmt, handler)
	handler = kaoriLog.LogRequestHandler(cfg.Logger.Connection, handler)

	// We need to shut down gracefully when the user hits Ctrl-C.
	serv := NewServer(handler)
	signal.Notify(serv.sig, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGTERM)

	serv.wg.Add(1)
	go serv.Start()

	// Wait for the interrupt signal.
	s := <-serv.sig
	switch s {
	case syscall.SIGTERM:

		// Go for the program exit. Don't wait for the server to finish.
		fmt.Println(pid, "Received SIGTERM, exiting without waiting for the web server to shut down")
		return

	case syscall.SIGINT:

		// Stop the server gracefully.
		fmt.Println(pid, "Received SIGINT")

	case syscall.SIGUSR1:

		// Spawn a child process.
		fmt.Println(pid, "Received SIGUSR1")
		var args []string
		if len(os.Args) > 1 {
			args = os.Args[1:]
		}
		file, err := serv.listener.File()
		if err != nil {
			fmt.Printf("%d Listener did not return file, not forking: %s\n", pid, err)
		} else {
			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.ExtraFiles = []*os.File{file}
			cmd.Env = append(os.Environ(), "FORKED_SERVER=1")
			if err := cmd.Start(); err != nil {
				fmt.Printf("%d Fork did not succeed: %s\n", pid, err)
			}
			fmt.Printf("%d Started child process %d, waiting for its ready signal\n", pid, cmd.Process.Pid)

			// We have a child process. A SIGTERM means the child process is ready to
			// start its server.
			<-serv.sig
		}
	}

	// Force the server to shut down.
	fmt.Println(pid, "Shutting down web server and waiting for requests to finish...")
	defer fmt.Println(pid, "Requests have finished")
	if err := serv.server.Shutdown(context.Background()); err != nil {
		log.Println(fmt.Errorf("Shutdown failed: %s", err))
		return
	}
	serv.wg.Wait()

	return
}

func (s *Server) Start() {
	defer s.wg.Done()

	var (
		ln  net.Listener
		err error
	)

	pid := os.Getpid()
	address := "0.0.0.0" + cfg.Server.Port

	// If this is a forked child process, we'll use its connection.
	isFork := os.Getenv("FORKED_SERVER") != ""

	if isFork {

		// It's a fork. Get the file that was handed over.
		kaoriLog.PrintLog("Server", "", "Start", fmt.Sprintf("%d Getting existing listener for %s\n", pid, address), 0)
		file := os.NewFile(3, "")
		ln, err = net.FileListener(file)
		if err != nil {
			kaoriLog.PrintLog("Server", "", "Start", fmt.Sprintf("%d Cannot use existing listener: %s\n", pid, err), 1)
			s.sig <- syscall.SIGTERM
			return
		}

		// Tell the parent to stop the server now.
		parent := syscall.Getppid()

		kaoriLog.PrintLog("Server", "", "Start", fmt.Sprintf("%d Telling parent process (%d) to stop server\n", pid, parent), 0)
		syscall.Kill(parent, syscall.SIGTERM)

		// Give the parent some time.
		time.Sleep(100 * time.Millisecond)

	} else {

		// It's a new server.
		kaoriLog.PrintLog("Server", "", "Start", fmt.Sprintf("%d Starting web server on %s\n", pid, address), 0)
		ln, err = net.Listen("tcp", address)
		if err != nil {
			kaoriLog.PrintLog("Server", "", "Start", fmt.Sprintf("%d Cannot listen to %s: %s\n", pid, address, err), 1)
			s.sig <- syscall.SIGTERM
			return
		}

	}

	// We can start the server now.
	kaoriLog.PrintLog("Server", "", "Start", 	fmt.Sprint(pid, " Serving requests..."), 0)

	s.listener = ln.(*net.TCPListener)

	err = s.server.ServeTLS(tcpKeepAliveListener{s.listener}, cfg.Server.Ssl.Certificate, cfg.Server.Ssl.Key)
	if err != nil {
		kaoriLog.PrintLog("Server", "", "Start", 	fmt.Sprintf("%d Web server was shut down: %s\n", pid, err), 2)
	}

	kaoriLog.PrintLog("Server", "", "Start", 	fmt.Sprint(pid, "Web server has finished"), 0)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		ip := kaoriUtils.GetIP(r)

		content, err := os.ReadFile(filepath.ToSlash(filepath.Join(cfg.Server.Gui, indexFile)))
		if err != nil {
			kaoriLog.PrintLog("General", ip, "ServeHttp", "Error to open index file", 1)
			kaoriUtils.PrintInternalErr(w)
			return
		}

		w.Write(content)

	} else {
		kaoriUtils.PrintErr(w, "")
		return
	}
}

func serveLogin(w http.ResponseWriter, r *http.Request) {

	ip := kaoriUtils.GetIP(r)

	data, err := os.ReadFile(filepath.Join(cfg.Server.Gui, loginFile))
	if err != nil {
		kaoriLog.PrintLog("General", ip, "serveLogin", "Error to open file: "+loginFile, 1)
		kaoriUtils.PrintInternalErr(w)
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
