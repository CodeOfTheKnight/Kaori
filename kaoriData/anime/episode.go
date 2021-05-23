package anime

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

type Episode struct {
	Number int `firestore:"number"`
	Title string
	Videos []*Video
}

func NewEpisode() *Episode {
	return &Episode{}
}

func (ep *Episode) CheckEpisode() error {

	if err := ep.checkNumber(); err != nil {
		return err
	}

	//TODO: Get title if it hasn't been set

	for i, _ := range ep.Videos {
		if err := ep.Videos[i].CheckVideo(); err != nil {
			return err
		}
	}

	return nil
}

func (ep *Episode) checkNumber() error {

	if ep.Number == 0 {
		return errors.New("Number of episode not setted")
	}

	if _, err := strconv.Atoi(strconv.Itoa(ep.Number)); err != nil {
		return errors.New("Number of episode not valid")
	}

	return nil
}

func (ep *Episode) SendToDbRel(cl *sql.DB, IdAnime int) (int, error) {

	//Insert AnimeInfo
	query := "INSERT INTO Episodi(Numero, Titolo, AnimeID) VALUES (?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	stmt, err := cl.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, ep.Number, ep.Title, IdAnime)
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
