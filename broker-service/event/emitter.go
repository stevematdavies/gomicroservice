package event

import (
	"github.com/rabbitmq/amqp091-go"
	"log"
)

type Emitter struct {
	connection *amqp091.Connection
}

func (e *Emitter) Handshake() error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return DeclareExchange(ch)
}

func (e *Emitter) Push(evt string, svy string) error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	log.Println("pushing to channel..")

	if err = ch.Publish("logs_topic", svy, false, false, amqp091.Publishing{
		ContentType: "text/plain",
		Body:        []byte(evt),
	}); err != nil {
		return err
	}
	return nil
}

func NewEventEmitter(c *amqp091.Connection) (Emitter, error) {
	e := Emitter{
		connection: c,
	}
	if err := e.Handshake(); err != nil {
		return Emitter{}, err
	}
	return e, nil
}
