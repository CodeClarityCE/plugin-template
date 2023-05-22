package template

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	types_amqp "codeclarity.io/types/amqp"
	types_database "codeclarity.io/types/database"

	amqp "github.com/rabbitmq/amqp091-go"
)

func receiveMessage(connection string) {
	// Create connexion
	host := os.Getenv("RABBITMQ_HOST")
	conn, err := amqp.Dial(host)
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	// Open channel
	ch, err := conn.Channel()
	if err != nil {
		failOnError(err, "Failed to open a channel")
	}
	defer ch.Close()

	// Declare queue
	q, err := ch.QueueDeclare(
		connection, // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
	}

	// Consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		failOnError(err, "Failed to register a consumer")
	}

	var forever chan struct{}
	go func() {
		for d := range msgs {
			// Start timer
			start := time.Now()

			dispatch(connection, d)

			// Print time elapsed
			t := time.Now()
			elapsed := t.Sub(start)
			log.Println(elapsed)
		}
	}()

	log.Printf(" [*] DISPATCHER Waiting for messages from " + connection + ". To exit press CTRL+C")
	<-forever
}

func dispatch(connection string, d amqp.Delivery) {
	if connection == "symfony_request" { // If message is from symfony_request
		// Read message
		var contextSymfony types_amqp.SymfonyDispatcherContext
		json.Unmarshal([]byte(d.Body), &contextSymfony)

		// Construct project document
		project := types_database.Project{
			Name: contextSymfony.Context.Project,
			UID:  contextSymfony.Context.Uid,
		}

		// Do something
		data, _ := json.Marshal(project)

		send("dispatcher_sbom", data)
	} else if connection == "sbom_dispatcher" { // If message is from sbom_dispatcher
		// Read message
		var context types_amqp.SbomDispatcherMessage
		err := json.Unmarshal([]byte(d.Body), &context)
		if err != nil {
			log.Println(err)
		}

		// Do something
		symfony_message := types_amqp.DispatcherSymfonyContext{
			Context: types_amqp.DispatcherSymfonyMessage{
				Analysis:  fmt.Sprintf("%d", 0),
				Analyzers: []string(nil),
				Sbom:      context.Sbom,
				Date: types_amqp.Date{
					Date:         "",
					TimezoneType: "",
				},
				Uid:  0,
				Step: "sbom",
			},
		}
		data, _ := json.Marshal(symfony_message)
		send("dispatcher_symfony", data)
	}

}
