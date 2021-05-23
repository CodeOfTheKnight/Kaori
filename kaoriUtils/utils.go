package kaoriUtils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/gorilla/securecookie"
	"github.com/hajimehoshi/go-mp3"
	"github.com/segmentio/ksuid"
	templateHtml "html/template"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type ParamsInfo struct {
	Key      string
	Required bool
}

const LittleBoxURI string = "https://litterbox.catbox.moe/resources/internals/api.php"
const UrlAnilist string = "https://anilist.co"

//lsGui ritorna la path e il nome dei file presenti nella cartella KaoriGui.
func lsGui(dir string) (files []string, err error) {
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
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

//ls ritorna la path e il nome dei file presenti in una directory.
func ls(dir string) (files []string, err error) {
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("{\"code\": 500, \"msg\": \"Internal Server Error\"}\n"))
}

//printErr ritorna un errore al client impostando a 400 lo status code della risposta HTTP.
func printErr(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("{\"code\": 400, \"msg\": \"%s\"}\n", err)))
}

//printOk ritorna scrive al client uno status code 200 per indicare che va tutto bene.
func printOk(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"code\": 200, \"msg\": \"OK\"}\n"))
}

//getParams ritorna i parametri inviati tramite metodo GET dell'HTTP request.
func getParams(params []ParamsInfo, r *http.Request) (map[string]interface{}, error) {

	values := make(map[string]interface{})

	for _, param := range params {

		fmt.Println(param)

		keys, err := r.URL.Query()[param.Key]

		fmt.Println(err)

		if (!err || len(keys[0]) < 1) && param.Required == true {
			return nil, errors.New(fmt.Sprintf("Url Param \"%s\" is missing", param.Key))
		}

		if err == false || len(keys[0]) < 1 {
			continue
		}

		values[param.Key] = keys[0]
	}

	return values, nil

}

//validateIdAnilist convalida l'id anilist, ritorna true se è corretto altrimenti false.
func validateIdAnilist(ida int, tipo string) bool {
	resp, _ := http.Get(fmt.Sprintf("%s/%s/%d/", UrlAnilist, tipo, ida))
	if resp.StatusCode == 200 {
		return true
	}

	return false
}

//Converts pre-existing base64 data (found in example of https://golang.org/pkg/image/#Decode) to test.png
func base64toPng(strcode string) ([]byte, error) {

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
	req, err := http.NewRequest("POST", LittleBoxURI, body)
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

func setCookies(w http.ResponseWriter, token, tokenIss, cookieKey string) error {

	var s = securecookie.New([]byte(cookieKey), nil)

	value := map[string]string{
		"RefreshToken": token,
	}
	if encoded, err := s.Encode(tokenIss, value); err == nil {
		cookie := &http.Cookie{
			Name:     tokenIss,
			Value:    encoded,
			Secure:   false,
			HttpOnly: false,
			Path:     "/",
		}
		fmt.Println("COOKIE", cookie)
		http.SetCookie(w, cookie)
	} else {
		return err
	}

	return nil
}

func getCookies(r *http.Request, tokenIss, cookieKey string) (map[string]string, error) {

	var s = securecookie.New([]byte(cookieKey), nil)

	if cookie, err := r.Cookie(tokenIss); err == nil {
		value := make(map[string]string)
		if err = s.Decode(tokenIss, cookie.Value, &value); err == nil {
			return value, nil
		}
	}

	return nil, errors.New("Cookies not valid!")
}

func verifyAuth(db *kaoriDatabase.NoSqlDb, email, password string) (bool, error) {

	document, err := db.Client.C.Collection("User").Doc(email).Get(db.Client.Ctx)
	if err != nil {
		return false, err
	}

	data := document.Data()

	active := data["IsActive"].(bool)

	if !active {
		return false, errors.New("inactive")
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

func parseTemplateHtml(tmpl string, data interface{}) (string, error) {

	templatePath, err := filepath.Abs(tmpl)
	if err != nil {
		return "", errors.New("invalid template name")
	}

	t, err := templateHtml.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	if err = t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func existUser(db *kaoriDatabase.NoSqlDb, email string) bool {
	_, err := db.Client.C.Collection("User").Doc(email).Get(db.Client.Ctx)
	if err != nil {
		return false
	} else {
		return true
	}
}

func passwordValid(pws string) error {
	if len(pws) < 8 {
		return fmt.Errorf("password len is < 9")
	}
	if b, err := regexp.MatchString(`[0-9]{1}`, pws); !b || err != nil {
		return fmt.Errorf("password need num")
	}
	if b, err := regexp.MatchString(`[a-z]{1}`, pws); !b || err != nil {
		return fmt.Errorf("password need a_z")
	}
	if b, err := regexp.MatchString(`[A-Z]{1}`, pws); !b || err != nil {
		return fmt.Errorf("password need A_Z")
	}
	return nil
}

func portValid(port string) error {
	portInt, err := strconv.Atoi(strings.Trim(port, ":"))
	if err != nil {
		return errors.New("Invalid Port: Conversion of port to int not valid")
	}

	if portInt < 1024 || portInt > 49151 {
		return errors.New("Port not valid. [1024-49151]")
	}
	return nil
}

func checkHash(hash string) bool {
	ok, _ := regexp.MatchString(`^#(?:[0-9a-fA-F]{3}){1,2}$`, hash)
	return ok
}

func filterLog(rows []string, filter string, filterValue string) (r []string, err error) {

	for _, row := range rows {

		var objmap map[string]json.RawMessage
		err = json.Unmarshal([]byte(row), &objmap)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		str, err := objmap[filter].MarshalJSON()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		switch filter {
		case "msg":
			if strings.Contains(strings.Trim(string(str), "\""), filterValue) {
				r = append(r, row)
			}
		case "time":

			if strings.Contains(filterValue, "-") {
				dateMatrix := strings.Split(filterValue, "-")

				i, err := strconv.ParseInt(dateMatrix[0], 10, 64)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
				date1 := time.Unix(i, 0)

				i, err = strconv.ParseInt(dateMatrix[1], 10, 64)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
				date2 := time.Unix(i, 0)

				dateLog, err := time.Parse(time.RFC3339, strings.Trim(string(str), "\""))
				if err != nil {
					fmt.Println(err)
					return nil, err
				}

				if (dateLog.Unix() >= date1.Unix()) && (dateLog.Unix() <= date2.Unix()) {
					r = append(r, row)
				}
			} else {

				i, err := strconv.ParseInt(filterValue, 10, 64)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
				date := time.Unix(i, 0)

				dateLog, err := time.Parse(time.RFC3339, strings.Trim(string(str), "\""))
				if err != nil {
					fmt.Println(err)
					return nil, err
				}

				if dateLog.Unix() >= date.Unix() {
					r = append(r, row)
				}
			}

		default:
			if strings.Trim(string(str), "\"") == filterValue {
				r = append(r, row)
			}
		}

	}
	return r, err
}

func checkFiltersLogGet(params []ParamsInfo, values map[string]interface{}) error {

	keys := strings.Split(values["order"].(string), ",")

	myFunc := func(params []ParamsInfo, key string) bool {
		for _, param := range params {
			if param.Key == key {
				return true
			}
		}
		return false
	}

	for _, key := range keys {
		if !myFunc(params, key) {
			return errors.New("Params in order filed not valid!")
		}
	}

	return nil
}

func checkFiltersLogPost(values map[string]interface{}) error {
	var keys []string
	orders := strings.Split(values["order"].(string), ",")

	for key, _ := range values {
		keys = append(keys, key)
	}

	myFunc := func(keys []string, order string) bool {
		for _, key := range keys {
			if key == order {
				return true
			}
		}
		return false
	}

	for _, order := range orders {
		if !myFunc(keys, order) {
			return errors.New("Params in order filed not valid!")
		}
	}

	return nil
}

func hasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}
