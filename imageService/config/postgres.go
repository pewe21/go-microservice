package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() *PostgresStorage {
	connStr := "postgres://postgress:postgress@localhost/pqgotest?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Cannot establish connection to database: ", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot ping to database: ", err.Error())
	}

	log.Println("Connected to database")

	return &PostgresStorage{
		db: db,
	}
}
