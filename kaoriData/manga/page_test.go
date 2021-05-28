package manga

import (
	"database/sql"
	"testing"
)

func TestPage_SendToDB(t *testing.T)  {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriManga")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	pages, err := GetPageFromDB(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pages)

}
