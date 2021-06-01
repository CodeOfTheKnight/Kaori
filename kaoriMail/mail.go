package kaoriMail

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"github.com/CodeOfTheKnight/kaoriData"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

//Client for read and modify mail but not delete
type Service struct {
	*gmail.Service
	Mails map[string]Mail
}

const (
	MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

type Mail struct {
	Msg    string
	Labels string
	Head   []*gmail.MessagePartHeader
	Attach []*Attach
}

type Attach struct {
	FileName string
	Size     int64
	Mime     string
	Data     []byte
}

func ReadMailConfig(scope ...string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, scope...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}

	return config, nil
}

func GetClient(config *oauth2.Config) *http.Client {
	tokFile := "kaoriSrc/tokens/gmailToken.json"
	tok, err := TokenFromFile(tokFile)
	if err != nil {
		tok = GetTokenFromWeb(config)
		SaveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func SaveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func NewService(client *http.Client) (Service, error) {
	srv, err := gmail.New(client)
	if err != nil {
		return Service{}, errors.New(fmt.Sprintf("Unable to retrieve Gmail client: %v", err))
	}

	return Service{srv, make(map[string]Mail)}, nil
}

func (s *Service) GetMails() error {

	//Do request
	mreq := s.Users.Messages.List("me")
	mess, err := mreq.Do()
	if err != nil {
		return err
	}

	for _, m := range mess.Messages {

		var attachs []*Attach
		var msg string

		idm := m.Id

		//Do request for labels and snippets
		mreq2 := s.Users.Messages.Get("me", idm)
		m, err := mreq2.Do()
		if err != nil {
			return err
		}

		//Extract attachments
		parts := m.Payload.Parts
		for _, h := range parts {

			fmt.Println(h.MimeType)

			if h.MimeType == "text/html" {
				continue
			}

			if len(h.Parts) != 0 {
				reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(h.Parts[0].Body.Data))
				contain, err := ioutil.ReadAll(reader)
				if err != nil {
					return err
				}
				msg = string(contain)
				continue
			}

			if h.MimeType != "text/plain" {

				attachReq := s.Users.Messages.Attachments.Get("me", idm, h.Body.AttachmentId)
				attach, err := attachReq.Do()
				if err != nil {
					return err
				}

				decoded, err := base64.URLEncoding.DecodeString(attach.Data)
				if err != nil {
					return err
				}
				str := base64.StdEncoding.EncodeToString(decoded)

				attachs = append(attachs, &Attach{
					FileName: h.Filename,
					Size:     h.Body.Size,
					Mime:     h.MimeType,
					Data:     []byte(str),
				})

			} else {
				attachs = append(attachs, &Attach{
					FileName: h.Filename,
					Size:     h.Body.Size,
					Mime:     h.MimeType,
					Data:     []byte(h.Body.Data),
				})
			}

		}

		s.Mails[idm] = Mail{
			Msg:    msg,
			Labels: strings.Join(m.LabelIds, ";"),
			Head:   m.Payload.Headers,
			Attach: attachs,
		}
	}
	return nil
}

func (s *Service) UnreadMails() {
	for idm, mail := range s.Mails {
		if !strings.Contains(mail.Labels, "UNREAD") {
			delete(s.Mails, idm)
		}
	}
}

func (s *Service) AddDataMails() error {
	for idm, mail := range s.Mails {
		for _, h := range mail.Head {
			if h.Name == "Subject" {
				h.Value = strings.Replace(h.Value, " ", "", -1)
				h.Value = strings.ToUpper(h.Value)
				if strings.Contains(h.Value, "|") {
					matrix := strings.Split(h.Value, "|")
					if matrix[0] == "ADD_DATA" {
						if matrix[1] != "MUSIC" {
							return errors.New(`L'unica modalità di aggiunta dati per ora è solo per le musiche!
Aggiungere nell'oggetto della mail "ADD_DATA | Music"`)
						}
					} else {
						delete(s.Mails, idm)
					}
				} else {
					delete(s.Mails, idm)
				}
			}
		}
	}
	return nil
}

func (s *Service) SetMailAtRead(idm string) error {
	req2 := s.Users.Messages.Modify("me", idm, &gmail.ModifyMessageRequest{
		AddLabelIds:     nil,
		RemoveLabelIds:  []string{"UNREAD"},
		ForceSendFields: nil,
		NullFields:      nil,
	})
	_, err := req2.Do()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SendEmailOAUTH2(to string, sub string, data interface{}, tmpl string) (bool, error) {

	emailBody, err := kaoriUtils.ParseTemplate(tmpl, data)
	if err != nil {
		return false, errors.New("unable to parse email template")
	}

	var message gmail.Message

	emailTo := "To: " + to + "\r\n"
	subject := "Subject: " + sub + "\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(emailTo + subject + mime + "\n" + emailBody)

	message.Raw = base64.URLEncoding.EncodeToString(msg)

	// Send the message
	_, err = s.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Mail) ParseMailMusic() (*kaoriData.MusicData, error) {

	var mu kaoriData.MusicData

	err := m.CheckAllegati()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(m.Msg), &mu)
	if err != nil {
		return nil, errors.New("Dati inviati non validi!!\nLeggere attentamente la documentazione.\nNel caso in cui si è sicuri di aver inviato i dati correttamente contattare gli sviluppatori.")
	}

	for i, a := range m.Attach {
		if m.Attach[i].Mime == "image/jpeg" {
			mu.Cover = fmt.Sprintf("data:%s;base64,%s", "image/jpeg", string(a.Data))
		} else {
			mu.Track = fmt.Sprintf("data:%s;base64,%s", "audio/mpeg", string(a.Data))
		}
	}
	return &mu, nil
}

func (m *Mail) CheckAllegati() error {

	//Numero allegati
	if len(m.Attach) != 2 {
		return errors.New("Numero di allegati invalido.\nLeggere la documentazione relativo all'invio di dati per mail.\nÈ necessario l'invio di una cover formato \".jpg\" e di una track formato \".mp3\"")
	}

	//Check allegati
	if m.Attach[0].Mime == "image/jpeg" && m.Attach[1].Mime == "audio/mpeg" {
		return nil
	}

	if m.Attach[0].Mime == "audio/mpeg" && m.Attach[1].Mime == "image/jpeg" {
		return nil
	}

	return errors.New("Allegati invalidi.\nLeggere la documentazione relativa all'invio di dati per mail.\nÈ necessario l'invio di una cover formato \".jpg\" e di una track formato \".mp3\"")
}

func (m *Mail) IsUser(db *kaoriDatabase.NoSqlDb) bool {
	_, err := db.Client.C.Collection("User").Doc(m.Head[4].Value).Get(db.Client.Ctx)
	if err != nil {
		return false
	}
	return true
}

func SendEmail(serverAddress, mailAddress, mailKey, to, sub, tmpl string,  data interface{}) error {
	from := mailAddress
	pass := mailKey

	emailBody, err := kaoriUtils.ParseTemplateHtml(tmpl, data)
	if err != nil {
		return errors.New("unable to parse email template")
	}

	fmt.Println(emailBody)

	msg := "From: " + from + "\n" +
		"To: " + to + "\r\n" + MIME + "\r\n" +
		emailBody

	err = smtp.SendMail(serverAddress,
		smtp.PlainAuth("", from, pass, strings.Split(serverAddress, ":")[0]),
		from, []string{to}, []byte(msg))

	if err != nil {
		return errors.New(fmt.Sprintf("smtp error: %s", err))
	}

	return nil
}
