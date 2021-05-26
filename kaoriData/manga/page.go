package manga

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type Page struct {
	Number int
	Language string
	Server string
	Link string
}

func (p *Page) SendToDB(cl *sql.DB, idCapitolo int) error {

	//Insert AnimeInfo
	query := "INSERT INTO Pagina(Numero, Lingua, Server, Link, ChapterID) VALUES (?, ?, ?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancelfunc()

	stmt, err := cl.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, p.Number, p.Language, p.Server, p.Link, idCapitolo)
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return err
	}

	return nil
}

func GetPageFromDB(db *sql.DB, idChapter int) (pages []*Page, err error) {

	// Execute the query
	smtp, err := db.Prepare("SELECT Numero, Lingua, Server, Link FROM Pagina WHERE ChapterID = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(idChapter)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		var p Page

		// for each row, scan the result into our tag composite object
		err = results.Scan(&p.Number, &p.Language, &p.Server, &p.Link)
		if err != nil {
			return nil, err
		}

		// and then print out the tag's Name attribute
		log.Println(p)

		pages = append(pages, &p)
	}

	return pages, nil
}