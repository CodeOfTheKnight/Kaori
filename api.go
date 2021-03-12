package main

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/KiritoNya/anilist-go"
	"log"
	"net/http"
)

type ParamsInfo struct {
	Key string
	Required bool
}

func ApiInfo(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		params := []ParamsInfo {
			{Key: "id", Required: true},
		}

		ip := GetIP(r)

		values, err := getParams(params, r)
		if err != nil {
			logger.WithField("function", "ApiInfo").Error(fmt.Sprint("%s: %s", ip, err))
			w.WriteHeader(400)
			w.Write([]byte("Bad request!"))
			return
		}

		log.Println(values["id"])
	}

}
