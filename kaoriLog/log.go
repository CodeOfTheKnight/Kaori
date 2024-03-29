package kaoriLog

import (
	"encoding/json"
	"github.com/felixge/httpsnoop"
	logger "github.com/sirupsen/logrus"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"net/http"
	"os"
	"sync"
	"time"
)

type ServerLog struct {
	Func string
	Ip string
	Level string
	Msg string
	Time time.Time
	User string
}

type HTTPReqInfo struct {
	Method    string        `json:"method"`
	Url       string        `json:"url"`
	Referer   string        `json:"ref"`
	Ipaddr    string        `json:"ip"`
	Code      int           `json:"code"` //Response Code 200, 400 ecc.
	Size      int64         `json:"size"` //Numero byte della risposta
	Duration  time.Duration `json:"duration"`
	Data      int64         `json:"data"`
	UserAgent string        `json:"agent"`
	muLogHTTP sync.Mutex
}

func (ri *HTTPReqInfo) LogHTTPReq(filePath string) {
	ri.muLogHTTP.Lock()
	out, err := json.Marshal(ri)
	if err != nil {
		PrintLog("Server", "", "logHTTPReq", "Error with JSON: "+err.Error(), 1)
		return
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		PrintLog("Server", "", "logHTTPReq", "Error to open file " + filePath, 1)
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(string(out) + "\n"); err != nil {
		PrintLog("Server", "", "logHTTPReq", "Error to write JSON in the file: "+err.Error(), 1)
		panic(err)
	}

	ri.muLogHTTP.Unlock()
}

func LogRequestHandler(filePath string, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ri := &HTTPReqInfo{
			Method:    r.Method,
			Url:       r.URL.String(),
			Referer:   r.Header.Get("Referer"),
			UserAgent: r.Header.Get("User-Agent"),
			Data:      time.Now().Unix(),
		}

		ri.Ipaddr = kaoriUtils.GetIP(r)

		// this runs handler h and captures information about
		// HTTP request
		m := httpsnoop.CaptureMetrics(h, w, r)

		ri.Code = m.Code
		ri.Size = m.Written
		ri.Duration = m.Duration
		ri.LogHTTPReq(filePath)
	}
	return http.HandlerFunc(fn)
}

func PrintLog(user, ip, function, msg string, lvl int) {
	switch lvl {
	case 0: //Info type log
		logger.WithFields(logger.Fields{
			"user":     user,
			"ip":       ip,
			"func": function,
		}).Info(msg)
	case 1: //Error type log
		logger.WithFields(logger.Fields{
			"user":     user,
			"ip":       ip,
			"func": function,
		}).Error(msg)
	case 2:
		logger.WithFields(logger.Fields{
			"user":     user,
			"ip":       ip,
			"func": function,
		}).Warn(msg)
	}
}

func (srvLog *ServerLog) Filter(filter string){

}
