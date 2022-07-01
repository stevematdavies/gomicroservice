package main

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"listener/event"
	"log"
	"math"
	"os"
	"time"
)

func main() {
	rmqc, err := connect()
	if err != nil {
		fmt.Println("Could not connect", err)
		os.Exit(1)
	}
	defer rmqc.Close()

	log.Println("Listening and consuming messages...")
	consumer, err := event.NewConsumer(rmqc)
	if err != nil {
		panic(err)
	}
	if err := consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"}); err != nil {
		log.Println(err)
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
