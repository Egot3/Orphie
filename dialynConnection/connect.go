package diacon

import (
	"context"
	"fmt"
	"log"
	"orphie/dialynConnection/handlers"
	"orphie/dialynConnection/queues"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(port string) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%v/", port))
	handlers.AMQPErrorHandler(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	handlers.AMQPErrorHandler(err, "Failed to open a chan")
	defer ch.Close()

	queues.NewTestQueue(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testQueue := queues.NewTestQueue(ch)

	body := "DINGRINGDINGRING"
	err = ch.PublishWithContext(ctx,
		"",
		testQueue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	handlers.AMQPErrorHandler(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
}
