package main

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	Email          string `json:"email"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	permission     string //[u]ser,[c]reator,[t]ester,[a]admin
	ProfilePicture string `json:"profilePicture,omitempty"`
	isDonator      bool
	isActive       bool
	anilistId      int //-1 se anilist non è stato collegato.
	dateSignUp     int64
	itemAdded      int //Numero di item aggiunti al database
	credits        int //Punti utili per guardare anime. Si guadagnano guardando pubblicità, donando o aggiungendo item al database.
	level          int //Si incrementa in base ai minuti passati sull'applicazione.
	badges         []Badge
	settings       Settings
	notifications  Notifications
	refreshToken   map[string]int64
}

type Settings struct {
	Graphics      GraphicSettings
	ShowBadge     bool
	IsPervert     bool
	ShowListAnime bool
	ShowListManga bool
}

type GraphicSettings struct {
	ThemePrimary       string
	ThemeSecondary     string
	ThemeFontPrimary   string
	ThemeFontSecondary string
	ThemeShadow        string
	Error              string
	FontError          string
}

func (u *User) NewUser() {
	u.permission = "u"
	u.isDonator = false
	u.isActive = false
	u.anilistId = -1 //Non connesso ad anilist
	u.dateSignUp = time.Now().Unix()
	u.itemAdded = 0
	u.credits = 20
	u.level = 1
	u.badges = []Badge{}
	u.settings = Settings{
		Graphics: GraphicSettings{
			ThemePrimary:       "",
			ThemeSecondary:     "",
			ThemeFontPrimary:   "",
			ThemeFontSecondary: "",
			ThemeShadow:        "",
			Error:              "",
			FontError:          "",
		},
		ShowBadge:     true,
		IsPervert:     false,
		ShowListAnime: true,
		ShowListManga: true,
	}
}

func (u *User) AddNewUser() error {

	err := kaoriUser.Client.AddUser(u)
	if err != nil {
		return err
	}

	return nil
}

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
