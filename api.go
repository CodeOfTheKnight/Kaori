package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type ParamsInfo struct {
	Key string
	Required bool
}

const jsonTestAddMusicFile = "tests/addData/music/addMusic.json"

func ApiAddMusic(w http.ResponseWriter, r *http.Request) {

		//TODO: Verificare che l'utente (se ha solo i permessi di utente) non abbia già caricato il massimo di canzoni (cioè 10)

		// Declare a new MusicData struct.
		var md MusicData

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&md)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = md.CheckError()
		if err != nil {
			log.Println(err)
			printErr(w, err.Error())
			return
		}

		md.GetNameAnime()
		fmt.Println(md.AnimeName)
		err = md.NormalizeName()
		if err != nil {
			log.Println(err)
			printInternalErr(w)
			return
		}

		//Upload
		err = md.UploadTemporaryFile()
		if err != nil {
			printInternalErr(w)
			return
		}

		//TODO: Verificare che non sia già nel database

		//Add to database
		err = md.AddDataToTmpDatabase()
		if err != nil {
			log.Println(err)
			printInternalErr(w)
			return
		}
}

//TESTS
//TODO: Inserire l'autenticazione di amministratore

func TestAddDataMusicJson(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		w.Write([]byte("CIAONE"))
		return
	}

	if r.Method == http.MethodGet {


		content, err := os.ReadFile(jsonTestAddMusicFile)
		if err != nil {
			log.Println(err)
			printInternalErr(w)
			return
		}

		w.Write(content)
	}
}
