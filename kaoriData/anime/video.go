package anime

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Video struct {
	Language string
	Modality string
	Quality *InfoQuality
	Server string
	StreamLink *StreamLink
}

type InfoQuality struct {
	Width int `firestore:"width"`
	Height int `firestore:"height"`
}

type StreamLink struct{
	Link string `firestore:"link"`
	Fansub string `firestore:"fansub"`
	Duration float64 `firestore:"duration"`
	Bitrate int `firestore:"bitrate"`
}

func NewVideo() *Video {
	var sl StreamLink
	var i InfoQuality
	return &Video{StreamLink: &sl, Quality: &i}
}

func (v *Video) CheckVideo() error {

	if err := v.checkLanguage(); err != nil {
		return err
	}

	if err := v.checkStreamLinks(); err != nil {
		return err
	}

	if err := v.checkQuality(); err != nil {
		return err
	}

	if err := v.checkServer(); err != nil {
		return err
	}

	return nil
}

func (v *Video) GetQuality(link string)  error {

	command := fmt.Sprintf("ffprobe -v error -select_streams v:0 -show_entries stream=width,height,duration,bit_rate -of default=noprint_wrappers=1 %s", link)

	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		v.Quality.Height = 0
		v.Quality.Width = 0
		return err
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Split(line, "=")
		switch(fields[0]){
		case "width":
			num, _ := strconv.Atoi(fields[1])
			v.Quality.Width = num
		case "height":
			num, _ := strconv.Atoi(fields[1])
			v.Quality.Height = num
		case "duration":
			num, _ := strconv.ParseFloat(fields[1], 64)
			v.StreamLink.Duration = num
		case "bit_rate":
			num, _ := strconv.Atoi(fields[1])
			v.StreamLink.Bitrate = num
		}
	}

	return nil
}

func (v *Video) checkLanguage() error {

	if v.Language == "" {
		return errors.New("Language not setted")
	}

	if v.Language != "SubIta" && v.Language != "Ita" {
		return errors.New("Language not valid. Only languages: \"SubIta\", \"Ita\" is valid")
	}

	return nil
}

func (v *Video) checkQuality() error {

	if v.Quality.Width < 0 {
		return errors.New("Width not valid")
	}

	if v.Quality.Height < 0 {
		return errors.New("Height not valid")
	}

	if v.Quality.Height == 0 && v.Quality.Width == 0 {
		_ = v.GetQuality(v.StreamLink.Link)
	}

	return nil
}

func (v *Video) checkStreamLinks() error {

	resp, _ := http.Get(v.StreamLink.Link)
	if resp.StatusCode != http.StatusOK {
		return errors.New("Stream link not valid")
	}
	resp.Body.Close()


	if v.StreamLink.Duration < 0 {
		return errors.New("Duration not valid")
	}

	if v.StreamLink.Bitrate < 0 {
		return errors.New("Bitrate not valid")
	}

	//TODO: Fansub check with a file with all real fansub

	return nil
}

func (v *Video) checkServer() error {
	if v.Server == "" {
		return errors.New("Server not setted")
	}
	return nil
}

func (v *Video) SendToDbRel(cl *sql.DB, episodeID int) (int, error) {

	fmt.Println("Episode{ID}: ", episodeID)
	
	//Insert AnimeInfo
	query := "INSERT INTO Video(Lingua, Width, Height, Bitrate, Durata, Fansub, Server, Link, EpisodeID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancelfunc()

	stmt, err := cl.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, v.Language, v.Quality.Width, v.Quality.Height, v.StreamLink.Bitrate, v.StreamLink.Duration, v.StreamLink.Fansub, v.Server, v.StreamLink.Link, episodeID)
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return -1, err
	}

	prdID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error %s when getting last inserted product",     err)
		return -1, err
	}

	return int(prdID), nil
}

func GetVideoFromDB(db *sql.DB, episodeID int) (videos []*Video, err error) {


	// Execute the query
	smtp, err := db.Prepare("SELECT Lingua, Width, Height, Bitrate, Durata, Fansub, Server, Link FROM Video WHERE EpisodeID = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(episodeID)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		var v Video
		var q InfoQuality
		var s StreamLink

		// for each row, scan the result into our tag composite object
		err = results.Scan(
				&v.Language,
				&q.Width,
				&q.Height,
				&s.Bitrate,
				&s.Duration,
				&s.Fansub,
				&v.Server,
				&s.Link,
		)
		if err != nil {
			return nil, err
		}

		v.Quality = &q
		v.StreamLink = &s

		// and then print out the tag's Name attribute
		log.Println(v)

		videos = append(videos, &v)
	}

	return videos, nil
}

