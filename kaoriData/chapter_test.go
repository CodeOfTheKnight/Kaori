package manga

import (
	"database/sql"
	"testing"
)

func TestChapter_GetChapterFromDB(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriManga")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	mangas, err := GetChapterFromDB(db, 455)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mangas)

}
