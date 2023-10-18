package main

import (
	"context"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

/*
used in Kn workflow as an entrypoint for calling Python handler.
*/

type envConfig struct {
	Msg  string `envconfig:"MESSAGE" default:"boring default msg, change me with env[MESSAGE]"`
	Type string `envconfig:"TYPE"`
}

var (
	env envConfig
)

// Define a small struct representing the data for our expected data.
type payload struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func gotEvent(inputEvent event.Event) (*event.Event, protocol.Result) {
	data := &payload{}
	if err := inputEvent.DataAs(data); err != nil {
		log.Printf("Got error while unmarshalling data: %s", err.Error())
		return nil, http.NewResult(400, "got error while unmarshalling data: %w", err)
	}

	log.Println("Received a new event: ")
	log.Printf("[%v] %s %s: %+v", inputEvent.Time(), inputEvent.Source(), inputEvent.Type(), data)

	// append eventMsgAppender to message of the data
	data.Message = data.Message + env.Msg

	// Create output event
	outputEvent := inputEvent.Clone()

	if err := outputEvent.SetData(cloudevents.ApplicationJSON, data); err != nil {
		log.Printf("Got error while marshalling data: %s", err.Error())
		return nil, http.NewResult(500, "got error while marshalling data: %w", err)
	}

	// Resolve type
	if env.Type != "" {
		outputEvent.SetType(env.Type)
	}

	log.Println("Transform the event to: ")
	log.Printf("[%s] %s %s: %+v", outputEvent.Time(), outputEvent.Source(), outputEvent.Type(), data)

	return &outputEvent, nil
}

func main() {
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("listening on 8080, appending %q to events", env.Msg)
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(context.Background(), gotEvent))
}
