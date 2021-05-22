package kaoriDatabase

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
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

func NewNoSqlDb(projId, db string) (*database, error) {
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
