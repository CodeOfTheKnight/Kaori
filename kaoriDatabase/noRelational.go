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

func (d *NoSqlDb) UpdateField(coll string, doc string, path string, val interface{}) error {
	_, err := d.Client.C.Collection(coll).Doc(doc).Update (d.Client.Ctx, [] firestore.Update {{Path: path, Value: val}})
	if err != nil {
		return err
	}
	return nil
}

func (d *NoSqlDb) SetField(coll string, doc string, val interface{}, merge bool) (err error) {
	if merge {
		_, err = d.Client.C.Collection(coll).Doc(doc).Set(d.Client.Ctx, val, firestore.MergeAll)
	} else {
		_, err = d.Client.C.Collection(coll).Doc(doc).Set(d.Client.Ctx, val)
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *NoSqlDb) AppendArray(coll string, doc string, path string, val interface{}) error {

	dc := d.Client.C.Collection(coll).Doc(doc)

	_, err := dc.Update(d.Client.Ctx, []firestore.Update{
		{Path: path, Value: firestore.ArrayUnion(val)},
	})
	if err != nil {
		return err
	}

	return nil
}
