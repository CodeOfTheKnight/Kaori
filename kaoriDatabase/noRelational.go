package kaoriDatabase

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
)

type NoSqlDb struct {
	ProjectId string
	Database string
	Client ClientFirestore
}

type ClientFirestore struct {
	C *firestore.Client
	Ctx context.Context
}

func NewNoSqlDb(projId, db string) (*NoSqlDb, error) {
	var d NoSqlDb

	d.ProjectId = projId
	d.Database = db
	err := d.Connect()
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *NoSqlDb) Connect() error {
	d.Client.Ctx = context.Background()

	conf := &firebase.Config{ProjectID: d.ProjectId}
	app, err := firebase.NewApp(d.Client.Ctx, conf, option.WithCredentialsFile(d.Database))
	if err != nil {
		return errors.New(fmt.Sprintf("error initializing app: %v\n", err.Error()))
	}

	client, err := app.Firestore(d.Client.Ctx)
	if err != nil {
		return err
	}
	d.Client.C = client
	return nil
}
