package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Env struct {
	db *sql.DB
}

var Envdb *Env

func Init() {
	home, err := os.UserHomeDir()
	content, err := ioutil.ReadFile(filepath.Join(home, "server.prp"))
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(content), "\n")
	connString := lines[0]
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	err = db.Ping()

	if err != nil {
		panic(err)
	}
	Envdb = &Env{db: db}
}
