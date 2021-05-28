package manga

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestManga_SendToDatabase(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriManga")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	m := &Manga{
			Id:             455,
			Name:           "Citrus",
			ChaptersNumber: 10, 
			Chapters:       []*Chapter{
				{
					Number: 1,
					Title:  "Capitolo 1",
					Pages:  []*Page{
						{
							Number:   1,
							Language: "Ita",
							Server:   "Mangaworld",
							Link:     "http://pagina.jpg",
						},
						{
							Number:   2,
							Language: "Ita",
							Server:   "Mangaworld",
							Link:     "http://pagina2.jpg",
						},
					},
				},
				{
					Number: 2,
					Title:  "Capitolo 2",
					Pages:  []*Page{
						{
							Number:   1,
							Language: "Ita",
							Server:   "Mangaworld",
							Link:     "http://pagina.jpg",
						},
					},
				},
			},
	}
	
	err = m.SendToDatabase(db)
	if err != nil {
		t.Fatal(err)
	}
	
	t.Log("[OK]")
}

func TestGetMangaFromDB(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriManga")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	manga, err := GetMangaFromDB(db, 455)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(manga)
}