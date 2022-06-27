package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type LogEntry struct {
	ID		  string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type Models struct {
	LogEntry LogEntry
}

func New(c *mongo.Client) Models {
	client = c

	return Models{
		LogEntry: LogEntry{},
	}
}

func (l *LogEntry) Insert(e LogEntry) error {
	c := client.Database("logs").Collection("logs")
	_, err := c.InsertOne(context.TODO(),LogEntry{
		Name: e.Name,
		Data: e.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error inserting in to logs: ", err)
		return err
	}	
	return nil
}

func (l *LogEntry) All()([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
	defer cancel()


	c := client.Database("logs").Collection("logs")
	o := options.Find()
	o.SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}})
	
	x, err := c.Find(context.TODO(), bson.D{}, o)
	if err != nil {
		log.Println("Error finding all docs", err)
		return nil, err
	}

	defer x.Close(ctx)

	var logs []*LogEntry

	for x.Next(ctx) {
		var item LogEntry
		if err := x.Decode(&item); err != nil {
			log.Println("Error decoding log", err)
			return nil,err
		}
		logs = append(logs, &item)
	}
	return logs, nil
}