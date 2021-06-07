package anime

import (
	"database/sql"
	"testing"
)

func TestAnime_SendToDbRel(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	a := &Anime{
		Id:       344,
		Name:     "Citrus",
		Episodes: []*Episode{
			{
				Number: "1",
				Title: "L'incontro delle ragazze",
				Videos: []*Video{
					{
						Language:   "SubIta",
						Modality:   "",
						Quality:    &InfoQuality{
							Width:  1920,
							Height: 1080,
						},
						Server:     "AnimeWorld",
						StreamLink: &StreamLink{
							Link:     "https://animeworld.it/prova.mp4",
							Fansub:   "YURI",
							Duration: 1254.123,
							Bitrate:  12547,
						},

					},
					{
						Language:   "SubIta",
						Modality:   "",
						Quality:    &InfoQuality{
							Width:  1920,
							Height: 1080,
						},
						Server:     "AnimeWorld",
						StreamLink: &StreamLink{
							Link:     "https://animeworld.it/prova.mp4",
							Fansub:   "YURI",
							Duration: 1254.123,
							Bitrate:  12547,
						},
					},
				},
			},
			{
				Number: "2",
				Title: "La scopata delle due",
				Videos: []*Video{
					{
						Language:   "SubIta",
						Modality:   "",
						Quality:    &InfoQuality{
							Width:  1920,
							Height: 1080,
						},
						Server:     "AnimeWorld",
						StreamLink: &StreamLink{
							Link:     "https://animeworld.it/prova2.mp4",
							Fansub:   "YURI",
							Duration: 1254.123,
							Bitrate:  12547,
						},
					},
				},
			},
		},
	}

	err = a.SendToDbRel(db)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")
}

/*func TestGetAnimeFromDB(t *testing.T) {

	db, err := sql.Open("mysql", "root:Goghetto1106@tcp(192.168.1.4:3306)/KaoriAnime")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	a, err := GetAnimeFromDB(db, 335)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(a)
}*/
