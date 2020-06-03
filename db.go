package main

import (
	"io/ioutil"
	"log"

	"github.com/vinhjaxt/go-mysql-lib"
)

var db *mysql.DB

func init() {
	b, err := ioutil.ReadFile("./db_config.json")
	if err != nil {
		log.Panicln(err)
	}
	dbCfg := &mysql.Config{}
	err = json.Unmarshal(b, dbCfg)
	if err != nil {
		log.Panicln(err)
	}
	db, err = mysql.New(dbCfg)
	if err != nil {
		log.Panicln(err)
	}
}
