package anime

import (
	"cloud.google.com/go/firestore"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/CodeOfTheKnight/Kaori/kaoriData"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
	"github.com/fatih/structs"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/api/iterator"
	"log"
	"strconv"
	"time"
)

type Anime struct {
	Id int              `firestore:"-"`
	Name string         `firestore:"name"`
	Episodes []*Episode `firestore:"episodes"`
}

func NewAnime() *Anime {
	return &Anime{}
}

func (a *Anime) CheckAnime() error {

	if err := a.checkID(); err != nil {
		return err
	}

	//TODO: Check modality and change modality type string to type Modality
	for i, _ := range a.Episodes {
		if err := a.Episodes[i].CheckEpisode(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Anime) SendToDb(db *kaoriDatabase.NoSqlDb) error  {

	var eps []*Episode
	var l string

	//Episode info
	eps = a.Episodes

	//Write season info to database
	_, err := db.Client.C.Collection("Anime").
				Doc(strconv.Itoa(a.Id)).
				Set(db.Client.Ctx, map[string]string{
					"Name": a.Name,
				}, firestore.MergeAll)

	if err != nil {
		return err
	}

	//Write episodes of sesason to database
	for _, ep := range eps {

		for _, video :=  range ep.Videos {

			l = video.Language

			q := strconv.Itoa(video.Quality.Height) + "p"
			if q == "0p" {
				q = "undefined"
			}

			//Send streamLinks and create all collections
			_, err = db.Client.C.Collection("Anime").
				Doc(strconv.Itoa(a.Id)).
				Collection("Languages").
				Doc(video.Language).
				Collection("Episodes").
				Doc(strconv.Itoa(ep.Number)).
				Collection("Quality").
				Doc(q).
				Collection("Servers").
				Doc(video.Server).
				Set(db.Client.Ctx, structs.Map(video.StreamLink), firestore.MergeAll)

			if err != nil {
				return err
			}

			//Send language info
			_, err = db.Client.C.Collection("Anime").
				Doc(strconv.Itoa(a.Id)).
				Collection("Languages").
				Doc(video.Language).Set(db.Client.Ctx, map[string]string{
				"Modality": video.Modality,
			}, firestore.MergeAll)


			if err != nil {
				return err
			}

			//Send quality info
			_, err = db.Client.C.Collection("Anime").
				Doc(strconv.Itoa(a.Id)).
				Collection("Languages").
				Doc(video.Language).
				Collection("Episodes").
				Doc(strconv.Itoa(ep.Number)).
				Collection("Quality").
				Doc(q).
				Set(db.Client.Ctx, structs.Map(video.Quality), firestore.MergeAll)

		}

		//Send episode data
		_, err = db.Client.C.Collection("Anime").
										Doc(strconv.Itoa(a.Id)).
										Collection("Languages").
										Doc(l).
										Collection("Episodes").
										Doc(strconv.Itoa(ep.Number)).
										Set(db.Client.Ctx, map[string]string{
											"Title": ep.Title,
										}, firestore.MergeAll)

		if err != nil {
			return err
		}

	}

	return nil
}

func (a *Anime) SendToDbRel(cl *sql.DB) error {

	//Insert AnimeInfo
	query := "INSERT INTO Anime(ID, Name) VALUES (?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancelfunc()

	stmt, err := cl.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, a.Id, a.Name)
	if err != nil {
		log.Printf("Error %s when inserting row into products table", err)
		return err
	}

	for _, episode := range a.Episodes {

		idEp, err := episode.SendToDbRel(cl, a.Id)
		if err != nil {
			return err
		}

		for _, video := range episode.Videos {

			_, err := video.SendToDbRel(cl, idEp)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func (a *Anime) GetAnimeFromDb(c *firestore.Client, ctx context.Context) error {

	if a.Id == 0 {
		return errors.New("Id of anime not setted")
	}

	err := a.GetAnimeInfoFromDb(c, ctx)
	if err != nil {
		return err
	}

	err = a.GetAnimeEpisodeDb(c, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (a *Anime) GetAnimeInfoFromDb(c *firestore.Client, ctx context.Context) error {

	if a.Id == 0 {
		return errors.New("Id of anime not setted.")
	}

	//Get anime season info
	data, err := c.
		Collection("Anime").
		Doc(strconv.Itoa(a.Id)).
		Get(ctx)

	if err != nil {
		return errors.New(fmt.Sprintf("Error to get anime %d from database: %s", a.Id, err.Error()))
	}

	err = data.DataTo(a)
	if err != nil {
		return errors.New(fmt.Sprintf("Error to convert anime %d to anime struct: %s", a.Id, err.Error()))
	}

	fmt.Println("ANIME:", a)

	return nil
}

func (a *Anime) GetAnimeEpisodeDb(c *firestore.Client, ctx context.Context) error {

	//Take all languages
	iterLang := c.Collection("Anime").
				  Doc(strconv.Itoa(a.Id)).
				  Collection("Languages").
				  Documents(ctx)
	defer iterLang.Stop()

	for {

		docLanguage, err := iterLang.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.New(fmt.Sprintf("Error to get episode with anime id %d: %s", a.Id, err.Error()))
		}

		fmt.Println("LANG:", docLanguage.Ref.ID)

		iterEpisode := c.Collection("Anime").
			Doc(strconv.Itoa(a.Id)).
			Collection("Languages").
			Doc(docLanguage.Ref.ID).
			Collection("Episodes").
			Documents(ctx)
		defer iterEpisode.Stop()

		for {

			var ep Episode

			docEpisode, err := iterEpisode.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return errors.New(fmt.Sprintf("Error to get episode with anime id %d: %s", a.Id, err.Error()))
			}

			err = docEpisode.DataTo(&ep)
			if err != nil {
				return err
			}

			fmt.Println("EP:", docEpisode.Ref.ID)

			ep.Number, err = strconv.Atoi(docEpisode.Ref.ID)
			if err != nil {
				return err
			}

			//Get quality
			iterQuality := c.Collection("Anime").
					Doc(strconv.Itoa(a.Id)).
					Collection("Languages").
					Doc(docLanguage.Ref.ID).
					Collection("Episodes").
					Doc(strconv.Itoa(ep.Number)).
					Collection("Quality").
					Documents(ctx)
			defer iterQuality.Stop()

			for {

				docQuality, err := iterQuality.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return errors.New(fmt.Sprintf("Error to get episode languages with anime id %d: %s", a.Id, err.Error()))
				}

				fmt.Println("Quality:", docQuality.Ref.ID)

				//Get servers
				iterServers := c.Collection("Anime").
						Doc(strconv.Itoa(a.Id)).
						Collection("Languages").
						Doc(docLanguage.Ref.ID).
						Collection("Episodes").
						Doc(strconv.Itoa(ep.Number)).
						Collection("Quality").
						Doc(docQuality.Ref.ID).
						Collection("Servers").
						Documents(ctx)

					for {

						var v Video
						var iq InfoQuality
						var stream StreamLink

						docServers, err := iterServers.Next()
						if err == iterator.Done {
							break
						}
						if err != nil {
							return errors.New(fmt.Sprintf("Error to get episode languages with anime id %d: %s", a.Id, err.Error()))
						}

						err = docServers.DataTo(&stream)
						if err != nil {
							return err
						}

						fmt.Println("Server:", docServers.Ref.ID)

						v.Modality = docLanguage.Data()["Modality"].(string)
						v.Language = docLanguage.Ref.ID
						iq.Width = int(docQuality.Data()["Width"].(int64))
						iq.Width = int(docQuality.Data()["Height"].(int64))
						v.Quality = &iq
						v.Server = docServers.Ref.ID
						v.StreamLink = &stream

						fmt.Println("VIDEO:", v)

						ep.Videos = append(ep.Videos, &v)
					}
				}

				fmt.Println("EPISODIO:", ep)

				a.Episodes = append(a.Episodes, &ep)
			}
		}

		return nil
}

func (a *Anime) SendToKaori(kaoriUrl, token string) error {
	return kaoriData.SendToKaori(a, kaoriUrl, token)
}

func (a *Anime) AppendFile(filePath string) error {
	return kaoriData.AppendFile(a, filePath)
}

func (a *Anime) checkID() error {

	if a.Id == 0 {
		return errors.New("Id not setted")
	}

	return nil
}
