package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "8083"
	grpcPort = "8500"
	rpcPort  = "8501"
	mongoURL = "mongodb://mongo:27017"
)

var client *mongo.Client

type Config struct {
}

func connectToMongo() (*mongo.Client, error) {
	co := options.Client().ApplyURI(mongoURL)
	co.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	c, err := mongo.Connect(context.TODO(), co)
	if err != nil {
		log.Println("Error Connecting to mongo client", err)
		return nil, err
	}
	return c, nil
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = &mongo.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()
}
