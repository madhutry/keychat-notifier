package main

import (
	"database/sql"
)

type Env struct {
	db *sql.DB
}

var Envdb *Env

func Init() {
	connString := GetDBUrl()
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
