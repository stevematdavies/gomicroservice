package main

import (
	"context"
	"fmt"
	"log"
	"logging/data"
	"time"
)

type RPCServer struct {
}

type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(p RPCPayload, res *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{Name: p.Name, Data: p.Data, CreatedAt: time.Now()})
	if err != nil {
		log.Println("error writing to mongo: ", err)
		return err
	}
	*res = fmt.Sprintf("payload handled via RPC: %s ", p.Name)
	return nil
}
