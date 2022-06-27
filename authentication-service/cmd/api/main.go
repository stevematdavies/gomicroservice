package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "8082"

var counts int = 64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service")

	conn := connectToDb()

	if conn == nil {
		log.Panic("Can't connect to Db")
	}
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func openDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDb() *sql.DB {
	dsn := os.Getenv("DSN")
	for {
		connection, err := openDb(dsn)
		if err != nil {
			log.Println("Database not ready")
			counts++
		} else {
			log.Println("Connected to Database")
			return connection
		}
		if counts > 10 {
			log.Println(err)
			return nil
		}
		log.Println("Backing off for a few seconds...")
		time.Sleep(time.Second * 5)
		continue
	}
}
