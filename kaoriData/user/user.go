package user

import (
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/CodeOfTheKnight/Kaori/kaoriUtils"
	"strings"
	"time"
)

//User è una struttura con tutti i dati dell'utente.
type User struct {
	Email          string `json:"email"`
	Username       string `json:"username"`
	Password       string           `json:"-"`
	ProfilePicture string           `json:"profilePicture"`
	Permission     string           `json:"permission"`//[u]ser,[c]reator,[t]ester,[a]admin
	IsDonator      bool             `json:"isDonator"`
	IsActive       bool             `json:"isActive"`
	AnilistId      int              `json:"anilistId"`//-1 se anilist non è stato collegato.
	DateSignUp     int64            `json:"dateSignUp"`
	ItemAdded      int              `json:"itemAdded" `//Numero di item aggiunti al database
	Credits        int              `json:"credits"` //Punti utili per guardare anime. Si guadagnano guardando pubblicità, donando o aggiungendo item al database.
	Level          int              `json:"level"` //Si incrementa in base ai minuti passati sull'applicazione.
	Badges         []Badge          `json:"badges"`
	Settings       Settings         `json:"settings"`
	Notifications  Notifications    `json:"notifications"`
	RefreshToken   map[string]int64 `json:"-"`
}

//Settings è una struttura con tutte le preferenze dell'utente.
type Settings struct {
	Graphics      GraphicSettings `json:"graphics,omitempty"`
	ShowBadge     bool            `json:"showBadge,omitempty"`
	IsPervert     bool            `json:"isPervert,omitempty"`
	ShowListAnime bool            `json:"showListAnime,omitempty"`
	ShowListManga bool            `json:"showListManga,omitempty"`
}

//GraphicSettings è una struttura con tutte le preferenze grafiche dell'utente.
type GraphicSettings struct {
	Background		   string `json:"background,omitempty"`
	OnBackground	   string `json:"onBackground,omitempty"`
	Surface1		   string `json:"surface1,omitempty"`
	Surface2		   string `json:"surface2,omitempty"`
	Surface3		   string `json:"surface3,omitempty"`
	Surface4		   string `json:"surface4,omitempty"`
	Surface6		   string `json:"surface6,omitempty"`
	Surface8		   string `json:"surface8,omitempty"`
	Surface12		   string `json:"surface12,omitempty"`
	Surface16		   string `json:"surface16,omitempty"`
	Surface24		   string `json:"surface24,omitempty"`
	OnSurface		   string `json:"onSurface,omitempty"`
	Primary   		   string `json:"primary,omitempty"`
	PrimaryDark		   string `json:"primaryDark,omitempty"`
	OnPrimary		   string `json:"onPrimary,omitempty"`
	Secondary          string `json:"secondary,omitempty"`
	SecondaryDark      string `json:"secondaryDark,omitempty"`
	OnSecondary        string `json:"onSecondary,omitempty"`
	Error			   string `json:"error,omitempty"`
	OnError            string `json:"onError,omitempty"`
}

//NewUser è il costruttore dell'oggetto User.
func (u *User) NewUser() {
	u.Permission = "u"
	u.IsDonator = false
	u.IsActive = false
	u.AnilistId = -1 //Non connesso ad anilist
	u.DateSignUp = time.Now().Unix()
	u.ItemAdded = 0
	u.Credits = 20
	u.Level = 1
	u.Badges = []Badge{}
	u.Settings = Settings{
		Graphics: GraphicSettings{
			Background:    "",
			OnBackground:  "",
			Surface1:      "",
			Surface2:      "",
			Surface3:      "",
			Surface4:      "",
			Surface6:      "",
			Surface8:      "",
			Surface12:     "",
			Surface16:     "",
			Surface24:     "",
			OnSurface:     "",
			Primary:       "",
			PrimaryDark:   "",
			OnPrimary:     "",
			Secondary:     "",
			SecondaryDark: "",
			OnSecondary:   "",
			Error:         "",
			OnError:       "",
		},
		ShowBadge:     true,
		IsPervert:     false,
		ShowListAnime: true,
		ShowListManga: true,
	}
}

//AddNewUser aggiunge un nuovo utente al database.
func (u *User) AddNewUser(db *kaoriDatabase.NoSqlDb) error {

	var tmp interface{}
	tmp = struct {
		Username       string
		Password       string
		Permission     string
		ProfilePicture string
		IsDonator      bool
		IsActive       bool
		AnilistId      int
		DateSignUp     int64
		ItemAdded      int
		Credits        int
		Level          int
		Badges         []string
		Settings       Settings
	}{
		Username: u.Username,
		Password: u.Password,
		Permission: u.Permission,
		ProfilePicture: u.ProfilePicture,
		IsDonator: u.IsDonator,
		IsActive: u.IsDonator,
		AnilistId: u.AnilistId,
		DateSignUp: u.DateSignUp,
		ItemAdded: u.ItemAdded,
		Credits: u.Credits,
		Level: u.Level,
		Badges: []string{},
		Settings: u.Settings,
	}

	_, err := db.Client.C.Collection("User").Doc(u.Email).Set(db.Client.Ctx, tmp)
	if err != nil {
		return err
	}

	return nil
}

//IsValid verifica che i dati utente inviati dal client in fase di registrazione siano corretti.
func (u *User) IsValid() error {

	//Check email
	if u.Email == "" || strings.Contains(u.Email, "@") == false {
		return errors.New("Email not valid")
	}
	if len(strings.Replace(u.Email, "@", "", -1)) < 3 {
		return errors.New("Lenght of email not valid")
	}

	//Check Username
	if u.Username == "" {
		return errors.New("Username not valid")
	}

	return nil
}

//IsValid verifica che i dati delle preferenze inviati dall'utente siano corretti.
func (s *Settings) IsValid() error {

	//Controllo le impostazioni grafiche
	if err := s.Graphics.IsValid(); err != nil {
		return err
	}

	return nil
}

//IsValid verifica che i dati delle preferenze grafiche inviati dall'utente siano corretti.
func (gs *GraphicSettings) IsValid() error {

	fmt.Println("PRIMARY", gs.Primary)

	if !kaoriUtils.CheckHash(gs.Background) {
		if gs.Background != "" {
			return errors.New("Graphics Settings background not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.OnBackground) {
		if gs.OnBackground != "" {
			return errors.New("Graphics Settings onBackground not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface1) {
		if gs.Surface1 != "" {
			return errors.New("Graphics Settings surface1 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface2) {
		if gs.Surface2 != "" {
			return errors.New("Graphics Settings surface2 not valid.")
		}

	}

	if !kaoriUtils.CheckHash(gs.Surface3) {
		if gs.Surface3 != "" {
			return errors.New("Graphics Settings surface3 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface4) {
		if gs.Surface4 != "" {
			return errors.New("Graphics Settings surface4 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface6) {
		if gs.Surface6 != "" {
			return errors.New("Graphics Settings surface6 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface8) {
		if gs.Surface8 != "" {
			return errors.New("Graphics Settings surface8 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface12) {
		if gs.Surface12 != "" {
			return errors.New("Graphics Settings surface12 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface16) {
		if gs.Surface16 != "" {
			return errors.New("Graphics Settings surface16 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Surface24) {
		if gs.Surface24 != "" {
			return errors.New("Graphics Settings surface24 not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.OnSurface) {
		if gs.OnSurface != "" {
			return errors.New("Graphics Settings onSurface not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Primary) {
		if gs.Primary != "" {
			return errors.New("Graphics Settings primary not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.PrimaryDark) {
		if gs.PrimaryDark != "" {
			return errors.New("Graphics Settings primaryDark not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.OnPrimary) {
		if gs.OnPrimary != "" {
			return errors.New("Graphics Settings onPrimary not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Secondary) {
		if gs.Secondary != "" {
			return errors.New("Graphics Settings secondary not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.SecondaryDark) {
		if gs.SecondaryDark != "" {
			return errors.New("Graphics Settings secondaryDark not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.OnSecondary) {
		if gs.OnSecondary != "" {
			return errors.New("Graphics Settings onSecondary not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.Error) {
		if gs.Error != "" {
			return errors.New("Graphics Settings error not valid.")
		}
	}

	if !kaoriUtils.CheckHash(gs.OnError) {
		if gs.OnError != "" {
			return errors.New("Graphics Settings onError not valid.")
		}
	}

	return nil
}
