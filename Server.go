package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/KiritoNya/animeworld_V2/animeworld"
	"github.com/CodeOfTheKnight/KaoriProject/anilist"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", inizio)
	http.HandleFunc("/search", search)
	http.HandleFunc("/episode", getEpisode)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}

//Restituisce la pagina html iniziale al client richiedente
func inizio(w http.ResponseWriter, r *http.Request) {

	enableCors(&w) //CORS POLICY

	if r.Method == "GET" {
		log.Println("Request GET from: " + GetIP(r))
		file, _ := os.Open("/media/kirito/Anime_3.0/Progammi/KaoriServer/test.html")
		cont, _ := ioutil.ReadAll(file)
		w.Write(cont)
	}
}

//Restituisce l'IP del client che ha effettuato la richiesta
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

//Funzione che abilita le Cors Policy
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
	(*w).Header().Set("Access-Control-Allow-Headers:", "*")
}

//Funzione che effettua la richiesta alle API di anilist
func search(w http.ResponseWriter, r *http.Request) {

	enableCors(&w) //CORS POLICY

	type Request struct {
		Nome     string `json:"nome"`
		Lingua   string `json:"lingua"`
		PageSize int    `json:"pageSize"`
	}

	type Response struct {
		Nome         string               `json:"nome"`
		UrlCopertina anilist.CoverImage `json:"urlImage"`
	}

	//VARIABILI
	var resp []Response
	var req Request

	if r.Method == "POST" {

		//Lettura richiesta
		ip := GetIP(r)
		log.Println("Request POST from: " + ip)

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(ip + " - " + err.Error())
		}

		log.Println(ip + " - " + string(reqBody))

		//Controllo validità JSON
		if json.Valid(reqBody) != true {
			log.Println("JSON non valido")
			w.Write([]byte("JSON non valido"))
		}

		//Riempie la struttura  Request con i dati ottenuti dal JSON
		err = json.Unmarshal(reqBody, &req)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(req)

		//Creazione query
		query := "query ($page: Int, $perPage: Int, $search: String, $type: MediaType) " +
			"{" + anilist.PageQueryAll +
			"media (search: $search, type: $type) { id idMal title { " + req.Lingua + " } coverImage { medium color } } } }"

		//Definizione variabili query
		variables := struct {
			Search  string `json:"search"`
			Type    string `json:"type"`
			Page    int    `json:"page"`
			PerPage int    `json:"perPage"`
		}{
			req.Nome,
			"ANIME",
			0,
			req.PageSize,
		}

		//Creazione oggetto
		a, err := anilist.New()
		if err != nil {
			log.Fatal(err)
		}

		//Ricerca
		p, err := a.Page(query, variables)
		if err != nil {
			log.Fatal(err)
		}

		//Creazione slice
		for i := 0; i < len(p.Media); i++ {
			switch req.Lingua {
			case "romaji":
				resp = append(resp, Response{Nome: p.Media[i].Title.Romaji, UrlCopertina: p.Media[i].CoverImage})
			case "english":
				resp = append(resp, Response{Nome: p.Media[i].Title.English, UrlCopertina: p.Media[i].CoverImage})
			case "native":
				resp = append(resp, Response{Nome: p.Media[i].Title.Native, UrlCopertina: p.Media[i].CoverImage})
			}
		}

		//Creazione JSON
		b, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("error:", err)
		}

		//Scrittura risultati sul writer
		w.Write(b)
	}
}

func getEpisode(w http.ResponseWriter, r *http.Request) {

    type Episodio struct {
		Num   string `json:"num"`
		Url   string `json:"url"`
	}

    var ep []Episodio

	enableCors(&w) //CORS POLICY

	if r.Method == "POST" {
        
        //Inizializzazione richiesta
		ip := GetIP(r)
		log.Println("Request POST from: " + ip)

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(ip + " - " + string(reqBody))

        //Ricerca stagione su animeworld
        err = animeworld2.Search(string(reqBody))
		if err != nil {
			log.Println(ip + " - " + "<Error> Season not found")
			w.Write([]byte("<Error> Season not found"))
		}
		log.Println(ip + " - " + "Search...Success")

        //Creazione mappa animeworld
		seasons := animeworld2.GetSeasonMap()
        log.Println(seasons)

		episodes, err := animeworld2.Episodi(seasons[string(reqBody)].Link)
		if err != nil {
			log.Println(ip + " - " + "<Error> VVVID episodes")
			w.Write([]byte("<Error> VVVID episodes"))
		}
		log.Println(ip + " - " + "Episodes...Success")

        //Creazione struttura Episodio
        for _, episode := range episodes {
            ep = append(ep, Episodio{Num: episode.Num, Url: episode.Url})
        }
    
        //Creazione JSON
		b, err := json.Marshal(ep)
		if err != nil {
			fmt.Println("error:", err)
		}
        
        //Invio dati
		w.Write([]byte(b))
	}
}
