package kaoriDatabase

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type SqlDb struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
	Db string `yaml:"db" json:"db"`
	Driver string `yaml:"driver" json:"driver"`
	client *sql.DB
}

func (sdb *SqlDb) Connect() (err error) {

	sdb.client, err = sql.Open(sdb.Driver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", sdb.Username, sdb.Password, sdb.Host, sdb.Port, sdb.Db))
	if err != nil {
		panic(err.Error())
	}

	return nil
}

func (sdb *SqlDb) Close() (err error) {
	return sdb.client.Close()
}