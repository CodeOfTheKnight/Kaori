package kaoriDatabase

import (
	"testing"
)

func TestNewSqlDb(t *testing.T) {

	_, err := NewSqlDb("root", "Goghetto1106", "192.168.1.4", "3306", "KaoriAnime", "mysql")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")

}

func TestSqlDb_Connect(t *testing.T) {

	db, err := NewSqlDb("root", "Goghetto1106", "192.168.1.4", "3306", "KaoriAnime", "mysql")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Connect()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")
}

func TestSqlDb_Close(t *testing.T) {

	db, err := NewSqlDb("root", "Goghetto1106", "192.168.1.4", "3306", "KaoriAnime", "mysql")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Connect()
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("[OK]")
}
