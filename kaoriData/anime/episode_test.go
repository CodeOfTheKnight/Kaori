package anime

import (
	"database/sql"
	"testing"
)

func TestEpisode_SendToDbRel(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ep := &Episode{
		Number: 1,
		Title:  "Episodio fan service",
		Videos: nil,
	}

	num, err := ep.SendToDbRel(db, 335)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(num)

}
