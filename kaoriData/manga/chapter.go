package manga

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type Chapter struct {
	Number int
	Title string
	Pages []*Page
}

func (c *Chapter) SendToDB(cl *sql.DB, mangaID int) error {

	//Insert AnimeInfo
	query := "INSERT INTO Capitoli(Numero, Nome, MangaID) VALUES (?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancelfunc()

	stmt, err := cl.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, c.Number, c.Title, mangaID)
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return err
	}

	prdID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error %s when getting last inserted product",     err)
		return err
	}

	for _, pagina := range c.Pages {

		err = pagina.SendToDB(cl, int(prdID))
		if err != nil {
			return err
		}

	}

	return nil
}

func GetChaptersFromDB(db *sql.DB, animeID int) (chapters []*Chapter, err error) {

	// Execute the query
	smtp, err := db.Prepare("SELECT ID, Numero, Nome FROM Capitoli WHERE MangaID = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(animeID)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		var ch Chapter
		var id int

		// for each row, scan the result into our tag composite object
		err = results.Scan(&id, &ch.Number, &ch.Title)
		if err != nil {
			return nil, err
		}

		// and then print out the tag's Name attribute
		log.Println(ch)

		ch.Pages, err = GetPageFromDB(db, id)
		if err != nil {
			return nil, err
		}

		chapters = append(chapters, &ch)
	}

	return chapters, nil
}

func GetChapterFromDB(db *sql.DB, animeID int, numPag int) (*Chapter, error) {

	var ch Chapter

	// Execute the query
	smtp, err := db.Prepare("SELECT ID, Numero, Nome FROM Capitoli WHERE MangaID = ? AND Numero = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(animeID, numPag)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		var id int

		// for each row, scan the result into our tag composite object
		err = results.Scan(&id, &ch.Number, &ch.Title)
		if err != nil {
			return nil, err
		}

		// and then print out the tag's Name attribute
		log.Println(ch)

		ch.Pages, err = GetPageFromDB(db, id)
		if err != nil {
			return nil, err
		}
	}

	return &ch, nil
}