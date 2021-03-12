package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ls(dir string) (files []string, err error ){
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir(){
				if strings.Contains(path, "/css/") || strings.Contains(path, "/js/") || strings.Contains(path, "/lib/") {
					files = append(files, path)
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return files, nil
}

//Funzione che abilita le Cors Policy
func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers:", "*")
}

//Restituisce l'IP del client che ha effettuato la richiesta
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func printInternalErr(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error!\n"))
}

func printErr(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 - Bad Request!" + err + "\n"))
}

func getParams(params []ParamsInfo, r *http.Request) (map[string]interface{}, error) {

	values := make(map[string]interface{})

	for _, param := range params {
		keys, err := r.URL.Query()[param.Key]

		if (!err || len(keys[0]) < 1) && param.Required == true {
			return nil , errors.New(fmt.Sprint("Url Param \"%s\" is missing", param.Key))
		}

		values[param.Key] = keys[0]
	}

	return values, nil

}