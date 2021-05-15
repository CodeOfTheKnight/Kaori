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

type database struct {
	ProjectId string
	Database string
	Client ClientFirestore
}

type ClientFirestore struct {
	c *firestore.Client
	ctx context.Context
}

var (
	kaoriTmp *database
	kaoriUser *database
	kaoriDataDB *database
)

func NewDatabase(projId, db string) (*database, error) {
	var d database

	d.ProjectId = projId
	d.Database = db
	err := d.Connect()
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *database) Connect() error {
	d.Client.ctx = context.Background()

	conf := &firebase.Config{ProjectID: d.ProjectId}
	app, err := firebase.NewApp(d.Client.ctx, conf, option.WithCredentialsFile(d.Database))
	if err != nil {
		return errors.New(fmt.Sprintf("error initializing app: %v\n", err.Error()))
	}

	client, err := app.Firestore(d.Client.ctx)
	if err != nil {
		return err
	}
	d.Client.c = client
	return nil
}

func (cl *ClientFirestore) UpdateField(coll string, doc string, path string, val interface{}) error {
	_, err := cl.c.Collection(coll).Doc(doc).Update (cl.ctx, [] firestore.Update {{Path: path, Value: val}})
	if err != nil {
		return err
	}
	return nil
}

func (cl *ClientFirestore) SetField(coll string, doc string, val interface{}, merge bool) (err error) {
	if merge {
		_, err = cl.c.Collection(coll).Doc(doc).Set(cl.ctx, val, firestore.MergeAll)
	} else {
		_, err = cl.c.Collection(coll).Doc(doc).Set(cl.ctx, val)
	}
	if err != nil {
		return err
	}
	return nil
}

func (cl *ClientFirestore) AppendArray(coll string, doc string, path string, val interface{}) error {

	dc := cl.c.Collection(coll).Doc(doc)

	_, err := dc.Update(cl.ctx, []firestore.Update{
		{Path: path, Value: firestore.ArrayUnion(val)},
	})
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientFirestore) AddMusicTemp(md *MusicData) error {

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

	_, err := cl.c.Collection(md.Type).Doc(strconv.Itoa(md.IdAnilist)).Set(cl.ctx, tmp, firestore.MergeAll)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientFirestore) AddUser(u *User) error {
	var tmp interface{}
	tmp = struct {
		Username string
		Password string
		Permission string
		ProfilePicture string
		IsDonator bool
		IsActive bool
		AnilistId int
		DateSignUp int64
		ItemAdded int
		Credits int
		Level int
		Badges []string
		Settings Settings
	}{
		Username: u.Username,
		Password: u.Password,
		Permission: u.Permission,
		ProfilePicture: u.ProfilePicture,
		IsDonator: u.IsDonator,
		IsActive: u.IsDonator,
		AnilistId: u.AnilistId,
		DateSignUp: u.DateSignUp,
		ItemAdded: u.ItemAdded,
		Credits: u.Credits,
		Level: u.Level,
		Badges: []string{},
		Settings: u.Settings,
	}

	_, err := cl.c.Collection("User").Doc(u.Email).Set(cl.ctx, tmp)
	if err != nil {
		return err
	}

	return nil
}