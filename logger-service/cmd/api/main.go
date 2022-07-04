package main

import (
	"context"
	"fmt"
	"log"
	"logging/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "8083"
	grpcPort = "8901"
	rpcPort  = ":5001"
	mongoURL = "mongodb://mongo:27017"
)

var client *mongo.Client

type Config struct {
	Models data.Models
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

func (app *Config) rpcListen() error {
	log.Println("Starting RPC Server on port: ", rpcPort)
	listen, err := net.Listen("tcp", rpcPort)
	if err != nil {
		return err
	}
	defer listen.Close()
	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			fmt.Println("error: ", err)
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	err = rpc.Register(new(RPCServer))
	go app.rpcListen()
	go app.gRPCListen()

	log.Println("Starting Logging service on port: ", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	if err = srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}

	// go app.serve()
}
