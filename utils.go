package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"text/template"
	"time"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/hajimehoshi/go-mp3"
	"github.com/segmentio/ksuid"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const littleBoxURI string = "https://litterbox.catbox.moe/resources/internals/api.php"

//ls ritorna la path e il nome dei file presenti in una directory.
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

//Restituisce l'IP del client che ha effettuato la richiesta.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

//printInternalErr imposta a 500 lo status code della risposta HTTP.
func printInternalErr(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error!\n"))
}

//printErr ritorna un errore al client impostando a 400 lo status code della risposta HTTP.
func printErr(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintln("400 - Bad Request!", err)))
}

//getParams ritorna i parametri inviati tramite metodo GET dell'HTTP request.
func getParams(params []ParamsInfo, r *http.Request) (map[string]interface{}, error) {

	values := make(map[string]interface{})

	for _, param := range params {
		keys, err := r.URL.Query()[param.Key]

		if (!err || len(keys[0]) < 1) && param.Required == true {
			return nil , errors.New(fmt.Sprintf("Url Param \"%s\" is missing", param.Key))
		}

		values[param.Key] = keys[0]
	}

	return values, nil

}

//validateIdAnilist convalida l'id anilist, ritorna true se è corretto altrimenti false.
func validateIdAnilist(ida int, tipo string) bool {
	resp, _ := http.Get(fmt.Sprintf("https://anilist.co/%s/%d/", tipo, ida))
	if resp.StatusCode == 200 {
		return true
	}

	return false
}

//Converts pre-existing base64 data (found in example of https://golang.org/pkg/image/#Decode) to test.png
func base64toPng(strcode string) ([]byte, error){

	var bfg bytes.Buffer

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(strcode))
	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	err = png.Encode(&bfg, m)
	if err != nil {
		return nil, err
	}

	return bfg.Bytes(), nil
}

//Given a base64 string of a JPEG, encodes it into an JPEG image test.jpg
func base64toJpg(data string) ([]byte, error) {

	var bfg bytes.Buffer

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	err = jpeg.Encode(&bfg, m, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, err
	}

	return bfg.Bytes(), nil
}

func base64toMp3(data string) ([]byte, error) {

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	_, err := mp3.NewDecoder(reader)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func uploadLittleBox(data []byte, nameFile string) (uri string, err error) {

	//Write multipart-data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("time", "72h")
	writer.WriteField("reqtype", "fileupload")
	part, err := writer.CreateFormFile("fileToUpload", nameFile)
	if err != nil {
		return "", err
	}
	part.Write(data)
	err = writer.Close()
	if err != nil {
		return "", err
	}

	//Make request
	req, err := http.NewRequest("POST", littleBoxURI, body)
	if err != nil {
		return "", err
	}
	//Set headers request
	req.Header.Set("Content-Type", writer.FormDataContentType())

	//Do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	//read response
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func GenerateID() string {
	return ksuid.New().String()
}

func setCookies(w http.ResponseWriter, token string) error {

	//TODO: METTERE NEL FILE DI CONFIGURAZIONE
	var hashKey = []byte("very-secret")
	var s = securecookie.New(hashKey, nil)

	value := map[string]string{
		"RefreshToken": token,
	}
	if encoded, err := s.Encode("KaoriStream", value); err == nil {
		cookie := &http.Cookie{
			Name:  "KaoriStream",
			Value: encoded,
			Secure: true,
			HttpOnly: true,
		}
		fmt.Println("COOKIE", cookie)
		http.SetCookie(w, cookie)
	} else {
		return err
	}

	return nil
}

func getCookies(r *http.Request) (map[string]string, error) {

	var hashKey = []byte("very-secret")
	var s = securecookie.New(hashKey, nil)

	if cookie, err := r.Cookie("KaoriStream"); err == nil {
		value := make(map[string]string)
		if err = s.Decode("KaoriStream", cookie.Value, &value); err == nil {
			return value, nil
		}
	}

	return nil, errors.New("Cookies not valid!")
}

func verifyAuth(email, password string) (bool, error) {

	document, err := kaoriUser.Client.GetItem("User", email)
	if err != nil {
		return false, err
	}

	data := document.Data()

	if !data["IsActive"].(bool) {
		return false, errors.New("unactive")
	}

	p := data["Password"].(string)
	if p == password {
		return true, nil
	}
	return false, nil
}

func dateToUnix(date string) int64 {
	t, _ := time.Parse(time.RFC3339Nano, date)
	return t.Unix()
}

func parseTemplate(tmpl string, data interface{}) (string, error) {

	templatePath, err := filepath.Abs(tmpl)
	if err != nil {
		return "", errors.New("invalid template name")
	}

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	if err = t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
