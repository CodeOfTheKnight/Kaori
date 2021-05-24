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