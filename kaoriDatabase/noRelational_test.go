package kaoriDatabase

import "testing"

func TestNewNoSqlDb(t *testing.T) {

	_, err := NewNoSqlDb("kaori-504c3", "../database/kaori-504c3-firebase-adminsdk-5apba-f66a21203e.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")
}

func TestNoSqlDb_Connect(t *testing.T) {

	db, err := NewNoSqlDb("kaori-504c3", "../database/kaori-504c3-firebase-adminsdk-5apba-f66a21203e.json")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Connect()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")
}


