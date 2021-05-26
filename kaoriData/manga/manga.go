package manga

import (
	"cloud.google.com/go/firestore"
	"context"
	"database/sql"
	"github.com/CodeOfTheKnight/Kaori/kaoriData"
	"github.com/fatih/structs"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

type Manga struct {
	Id int
	Name string
	ChaptersNumber int
	Chapters []*Chapter
}

func (m *Manga) SendToKaori(kaoriServer, token string) error {
	return kaoriData.SendToKaori(m, kaoriServer, token)
}

func (m *Manga) SendToDatabaseNR(c *firestore.Client, ctx context.Context) error {

	ma := structs.Map(m)
	delete(ma, "Chapters")

	//Send manga data
	mangaDoc := c.Collection("Manga").Doc(strconv.Itoa(m.Id))
	_, err := mangaDoc.Set(ctx, ma, firestore.MergeAll)
	if err != nil {
		return err
	}

	for _, ch := range m.Chapters {

		for _, p := range ch.Pages {

			mc := structs.Map(ch)
			delete(mc, "Pages")

			//Send chapters data
			chapterDoc := mangaDoc.Collection("Languages").
										Doc(p.Language).
										Collection("Chapters").
										Doc(strconv.Itoa(ch.Number))

			_, err = chapterDoc.Set(ctx, mc, firestore.MergeAll)
			if err != nil {
				return err
			}

			//Send pages
			pagesDoc := chapterDoc.Collection("Pages").
									Doc(strconv.Itoa(p.Number)).
									Collection("Servers").
									Doc(p.Server)

			_, err = pagesDoc.Set(ctx, map[string]string{
				"Link": p.Link,
			}, firestore.MergeAll)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func (m *Manga) SendToDatabase(cl *sql.DB) error {

	//Send manga info
	query := "INSERT INTO Manga(ID, Nome) VALUES (?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancelfunc()

	stmt, err := cl.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, m.Id, m.Name)
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return err
	}

	for _, chapter := range m.Chapters {

		err = chapter.SendToDB(cl, m.Id)
		if err != nil {
			return err
		}

	}

	return nil
}

func (m *Manga) AppendFile(filePath string) error {
	return kaoriData.AppendFile(m, filePath)
}

func GetMangaFromDB(db *sql.DB, idManga int) (*Manga, error) {

	var m Manga

	// Execute the query
	smtp, err := db.Prepare("SELECT Nome FROM Manga WHERE ID = ?")
	if err != nil {
		return nil, err
	}

	results, err := smtp.Query(idManga)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {

		// for each row, scan the result into our tag composite object
		err = results.Scan(&m.Name)
		if err != nil {
			return nil, err
		}

		m.Chapters, err = GetChapterFromDB(db, idManga)
		if err != nil {
			return nil, err
		}

		// and then print out the tag's Name attribute
		log.Println(m)
	}

	return &m, nil
}
