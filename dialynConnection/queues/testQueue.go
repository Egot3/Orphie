package queues

import (
	"newsgetter/dialynConnection/handlers"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewTestQueue(ch *amqp.Channel) amqp.Queue { //FACTORY MUST GROW
	q, err := ch.QueueDeclare(
		"YO PHONE IS LINGING",
		false,
		false,
		false,
		false,
		nil,
	)
	handlers.AMQPErrorHandler(err, "Failed to declare a queue")
	return q
}
