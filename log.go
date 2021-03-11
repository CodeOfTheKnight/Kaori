package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

const logConnection string = "connection.log.json"
const logServer string = "logServer.log.json"

type HTTPReqInfo struct {
	Method  string `json:"method"`
	Url     string `json:"url"`
	Referer string `json:"referer"`
	Ipaddr  string `json:"ipaddr"`
	Code int `json:"code"` //Response Code 200, 400 ecc.
	Size int64 `json:"size"` //Numero byte della risposta
	Duration  time.Duration `json:"duration"`
	Data      int64         `json:"data"`
	UserAgent string        `json:"userAgent"`
	muLogHTTP sync.Mutex
}

func (ri *HTTPReqInfo) logHTTPReq() {
	ri.muLogHTTP.Lock()
	out, err := json.MarshalIndent(ri, "", "  ")
	if err != nil {
		log.Println(err)
	}

	f, err := os.OpenFile(logConnection, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(string(out) + "\n"); err != nil {
		panic(err)
	}

	ri.muLogHTTP.Unlock()
}


