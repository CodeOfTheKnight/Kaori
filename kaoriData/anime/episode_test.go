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
		Number: "1",
		Title:  "Episodio fan service",
		Videos: nil,
	}

	num, err := ep.SendToDbRel(db, 335)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(num)

}

func TestGetEpisodesFromDB(t *testing.T) {

	db, err := sql.Open("mysql", "kiritony_KiritoNya:Goghetto1106@tcp(65.19.141.67:3306)/kiritony_KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	eps, err := GetEpisodesFromDB(db, 223)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(eps)
}

func TestGetEpisodeFromDB(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	eps, err := GetEpisodeFromDB(db, 335, 1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(eps)
}
