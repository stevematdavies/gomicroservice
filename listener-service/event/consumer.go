package event

import (
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

type Consumer struct {
	conn  *amqp091.Connection
	qname string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(cn *amqp091.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: cn,
	}

	if err := consumer.Handshake(); err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (c *Consumer) Handshake() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	return DeclareExchange(ch)
}

func (c *Consumer) Listen(topics []string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := DeclareRandomQueue(ch)
	if err != nil {
		return err
	}
	for _, s := range topics {
		if err = ch.QueueBind(q.Name, s, "logs_topic", false, nil); err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	alw := make(chan bool)
	go func() {
		for i := range messages {
			var payload Payload
			_ = json.Unmarshal(i.Body, &payload)
			go handlePayload(payload)
		}
	}()

	fmt.Printf("waiting for message [Exchange, Queue] [logs_topic, %s]", q.Name)
	<-alw

	return nil
}

func handlePayload(p Payload) {
	switch p.Name {
	case "log", "event":
		if err := LogEvent(p); err != nil {
			log.Println(err)
		}
	case "auth":
	default:
		if err := LogEvent(p); err != nil {
			log.Println(err)
		}
	}
}
