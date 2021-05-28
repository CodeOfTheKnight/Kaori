package anime

import (
	"database/sql"
	"testing"
)

func TestVideo_SendToDbRel(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	v := &Video{
		Language:   "SubIta",
		Modality:   "",
		Quality:    &InfoQuality{
			Width:  1920,
			Height: 1080,
		},
		Server:     "AnimeWorld",
		StreamLink: &StreamLink{
			Link:     "https://video.mp4",
			Fansub:   "JUPPI",
			Duration: 2568.1245,
			Bitrate:  12456,
		},
	}

	num, err := v.SendToDbRel(db, 7)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(num)

}

func TestVideo_GetVideoFromDB(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	videos, err := GetVideoFromDB(db, 7)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(videos)

}
