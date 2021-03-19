package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
	"strconv"
	"time"
)

type Client struct {
	client *firestore.Client
}

func (c *Client) Connect(ctx *context.Context) error {
	conf := &firebase.Config{ProjectID: "kaori-504c3"}
	app, err := firebase.NewApp(*ctx, conf, option.WithCredentialsFile("database/kaori-504c3-firebase-adminsdk-5apba-f66a21203e.json"))
	if err != nil {
		return errors.New(fmt.Sprintf("error initializing app: %v\n", err.Error()))
	}

	client, err := app.Firestore(*ctx)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *Client) AddMusicTemp(ctx *context.Context, md *MusicData) error {

	var tmp interface{}
	tmp = map[string]struct {
		Title string
		Cover string
		Track string
		Status string
		UserEmail string
		CreatorEmail string
		DateAdd int64
		DateView int64
		DateChecked int64
	}{
		strconv.Itoa(md.NumSong): {
			Title: md.normalName,
			Cover: md.Cover,
			Track: md.Track,
			Status: "unchecked",
			UserEmail: "prova@gmail.com",
			CreatorEmail: "creator@gmail.com",
			DateAdd: time.Now().Unix(),
			DateView: 0,
			DateChecked: 0,
		},
	}

	_, err := c.client.Collection(md.Type).Doc(strconv.Itoa(md.IdAnilist)).Set(*ctx, tmp, firestore.MergeAll)
	if err != nil {
		return err
	}

	return nil
}
