package main

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const webPort = "8081"

type Config struct {
	Rmq *amqp091.Connection
}

func main() {
	rc, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rc.Close()

	a := Config{
		Rmq: rc,
	}

	log.Printf("Starting Broker Service on port: %s\n", webPort)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: a.routes(),
	}

	err = s.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

func connect() (*amqp091.Connection, error) {
	var x int64
	var bko = 1 * time.Second
	var con *amqp091.Connection

	for {
		c, err := amqp091.Dial("amqp://guest:guest@rabbitmq:5672")
		if err != nil {
			fmt.Println("RabbitMQ not ready...")
			x++
		} else {
			log.Println("Connected to RabbitMQ")
			con = c
			break
		}

		if x > 5 {
			fmt.Printf("something went wrong %s\n", err)
			return nil, err
		}

		bko = time.Duration(math.Pow(float64(x), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(bko)
		continue
	}

	return con, nil

}
