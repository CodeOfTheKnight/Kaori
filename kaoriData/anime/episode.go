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
	Number string `firestore:"number"`
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

	if ep.Number == "" {
		return errors.New("Number of episode not setted")
	}

	if _, err := strconv.Atoi(ep.Number); err != nil {
		return errors.New("Number of episode not valid")
	}

	return nil
}

func (ep *Episode) SendToDbRel(cl *sql.DB, IdAnime int) (int, error) {

	//Insert AnimeInfo
	query := "INSERT IGNORE INTO Episodi(Numero, Titolo, AnimeID) VALUES (?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5 * time.Second)
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

	if prdID == 0 {

		// Execute the query
		smtp, err := cl.Prepare("SELECT ID FROM Episodi WHERE AnimeID = ? AND Numero = ?")
		if err != nil {
			return -1, err
		}

		results, err := smtp.Query(IdAnime, ep.Number)
		if err != nil {
			return -1, err
		}
		defer results.Close()

		for results.Next() {

			var id int

			// for each row, scan the result into our tag composite object
			err = results.Scan(&id)
			if err != nil {
				return -1, err
			}

			return id, nil
		}

	}

	return int(prdID), nil
}

func GetEpisodesFromDB(db *sql.DB, idAnime int) (eps []*Episode, err error) {

	// Execute the query
	smtp, err := db.Prepare("SELECT ID, Numero, Titolo FROM Episodi WHERE AnimeID = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(idAnime)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		var ep Episode
		var id int

		// for each row, scan the result into our tag composite object
		err = results.Scan(&id, &ep.Number, &ep.Title)
		if err != nil {
			return nil, err
		}

		ep.Videos, err = GetVideoFromDB(db, id)
		if err != nil {
			return nil, err
		}

		eps = append(eps, &ep)
	}

	return eps, nil
}

func GetEpisodeFromDB(db *sql.DB, idAnime int, numEp int) (*Episode, error) {

	var ep Episode

	// Execute the query
	smtp, err := db.Prepare("SELECT ID, Numero, Titolo FROM Episodi WHERE AnimeID = ? AND Numero = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(idAnime, numEp)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		var id int

		// for each row, scan the result into our tag composite object
		err = results.Scan(&id, &ep.Number, &ep.Title)
		if err != nil {
			return nil, err
		}

		ep.Videos, err = GetVideoFromDB(db, id)
		if err != nil {
			return nil, err
		}
	}

	return &ep, nil
}
