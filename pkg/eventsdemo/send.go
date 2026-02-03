package eventsdemo

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"

	"github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	"github.com/cloudevents/sdk-go/v2/client"
)

type PubSubClient struct {
	cl client.Client
}

func NewPubSubClient(ctx context.Context, accessToken, projectID, topicID string) (*PubSubClient, error) {
	p, err := pubsub.New(ctx,
		pubsub.WithProjectID(projectID),
		pubsub.WithTopicID(topicID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub protocol: %w", err)
	}

	// todo: auth with accessToken
	c, err := client.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudevents client: %w", err)
	}

	return &PubSubClient{cl: c}, nil
}

// Send publishes a CloudEvent to a Google Cloud Pub/Sub topic.
// payload can be any struct or map that can be marshaled to JSON.
func (p *PubSubClient) Send(ctx context.Context, eventSource, eventType string, payload any) error {
	// Create a new event.
	e := event.New()
	e.SetID(uuid.New().String())
	e.SetSource(eventSource)
	e.SetType(eventType)
	if err := e.SetData(event.ApplicationJSON, payload); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Send the event.
	log.Printf("sending event to topic")
	if err := p.cl.Send(ctx, e); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	log.Printf("event sent successfully")
	return nil
}
