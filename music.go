package main

import (
	"bytes"
	"errors"
	anilist "github.com/kaiserbh/anilistgo/anilist/query"
	"os"
	"strings"
	"text/template"
)

//MusicData struct
type MusicData struct {
	IdAnilist  int    `json:"idAnilist"`
	AnimeName  string `json:"-"`
	Type       string `json:"type"`
	NumSong    int    `json:"numSong"`
	IsFull     bool   `json:"isFull"`
	Artist     string `json:"artist,omitempty"`
	NameSong   string `json:"nameSong,omitempty"`
	Cover      string `json:"cover"`
	Track      string `json:"track"`
	imgCover   []byte
	track      []byte
	normalName string
}

//Music struct
type Music struct {
	OP         []Track
	ED         []Track
	SoundTrack []Track
}

//Track struct
type Track struct {
	Name         string
	Artist       string
	IdSoundCloud int
	Links        string
}

//CheckError esegue tutti i controlli per verificare che non siano stati inviati al server dati errati.
func (md *MusicData) CheckError() (err error) {

	//Check IdAnilist
	if !validateIdAnilist(md.IdAnilist, "anime") {
		return errors.New("idAnilist not valid")
	}

	//Check Type
	if strings.ToLower(md.Type) != "op" && strings.ToLower(md.Type) != "opening" && strings.ToLower(md.Type) != "ed" && strings.ToLower(md.Type) != "ending" && strings.ToLower(md.Type) != "ost" && strings.ToLower(md.Type) != "soundtrack" {
		return errors.New(`Accept only "OP, ED, OST, SoundTrack"`)
	}

	//Check NumSong
	if md.NumSong < 0 {
		return errors.New(`numSong not valid`)
	}

	//TODO: Controllare se l'artista esiste con qualche database o magari soundcloud

	//Check cover
	if md.Cover == "" {
		return errors.New(`Cover not valid`)
	}
	if md.CheckImage() != nil {
		return errors.New("Cover not valid")
	}

	//Controllo mp3
	if md.Track == "" {
		return errors.New("Track not valid")
	}
	if md.CheckTrack() != nil {
		return errors.New("Track not valid")
	}

	return nil
}

//CheckImage controlla se è un'immagine e se è nei formati gestibili dal server
func (md *MusicData) CheckImage() (err error) {

	switch strings.Split(md.Cover[5:], ";base64")[0] {
	case "image/png":

		md.imgCover, err = base64toPng(strings.Split(md.Cover, "base64,")[1])
		if err != nil {
			return err
		}

		//Check size
		if len(md.imgCover) == 0 {
			return errors.New("Cover not valid")
		}

	case "image/jpeg":

		md.imgCover, err = base64toJpg(strings.Split(md.Cover, "base64,")[1])
		if err != nil {
			return err
		}

		//Check size
		if len(md.imgCover) == 0 {
			return errors.New("Cover not valid")
		}

	default:
		return errors.New("Cover not valid")
	}

	return nil
}

//CheckTrack controlla se è un file mp3.
func (md *MusicData) CheckTrack() (err error) {
	switch strings.Split(md.Track[5:], ";base64")[0] {
	case "@file/mpeg":

		md.track, err = base64toMp3(strings.Split(md.Track, "base64,")[1])
		if err != nil {
			return err
		}

		//Check size
		if len(md.track) == 0 {
			return errors.New("Track not valid")
		}

	case "audio/mpeg":
		md.track, err = base64toMp3(strings.Split(md.Track, "base64,")[1])
		if err != nil {
			return err
		}

		//Check size
		if len(md.track) == 0 {
			return errors.New("Track not valid")
		}

	default:
		return errors.New("Track not valid")
	}

	return nil
}

//GetNameAnime setta mediante l'id di anilist il nome dell'anime
func (md *MusicData) GetNameAnime() {
	media := anilist.NewMedia()
	media.FilterByID(md.IdAnilist)
	md.AnimeName = media.Title.Romaji
}

//NormalizeName genera il nome del file audio.
func (md *MusicData) NormalizeName() error {

	var buf bytes.Buffer

	musicName, err := os.ReadFile(cfg.Template.Music["file"])

	if strings.ToLower(md.Type) == "ending" {
		md.Type = "ED"
	}
	if strings.ToLower(md.Type) == "opening" {
		md.Type = "OP"
	}
	if strings.ToLower(md.Type) == "soundtrack" {
		md.Type = "SoundTrack"
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("musicName").Parse(string(musicName)))

	err = t.Execute(&buf, md)
	if err != nil {
		return err
	}

	md.normalName = buf.String()

	return nil
}

//UploadTemporaryFile carica un file temporaneo su "littlebox" che scade dopo 3 giorni.
func (md *MusicData) UploadTemporaryFile() error {

	//UploadCover
	link, err := uploadLittleBox(md.imgCover, "cover.jpg")
	if err != nil {
		return err
	}
	md.Cover = link

	//UploadTrack
	md.Track, err = uploadLittleBox(md.track, md.normalName+".mp3")
	if err != nil {
		return err
	}

	return nil
}

//Aggiunge il dato al database
func (md *MusicData) AddDataToTmpDatabase() error {

	err := kaoriTmp.Client.AddMusicTemp(md)
	if err != nil {
		return err
	}

	return nil
}
