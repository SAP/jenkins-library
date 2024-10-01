package gcp

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"os"
)

type PubsubClient struct {
	projectID   string
	pool        string
	provider    string
	topic       string
	orderingKey string
}

func NewGcpPubsubClient(projectID, pool, provider, topic, orderingKey string) *PubsubClient {
	return &PubsubClient{
		projectID:   projectID,
		pool:        pool,
		provider:    provider,
		topic:       topic,
		orderingKey: orderingKey,
	}
}

func (cl *PubsubClient) Publish(data []byte) error {
	oidcToken := os.Getenv("PIPER_OIDCIdentityToken")
	accessToken, err := GetFederatedToken(cl.projectID, cl.pool, cl.provider, oidcToken)
	if err != nil {
		return errors.Wrap(err, "could not get federated token")
	}

	return publish(cl.projectID, accessToken, cl.topic, cl.orderingKey, data)
}

func publish(projectNumber, accessToken, topic, orderingKey string, data []byte) error {
	ctx := context.Background()

	staticTokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	pubsubClient, err := pubsub.NewClient(ctx, projectNumber, option.WithTokenSource(staticTokenSource))
	if err != nil {
		return errors.Wrap(err, "pubsub client creation failed")
	}

	msg := &pubsub.Message{Data: data, OrderingKey: orderingKey}
	publishResult := pubsubClient.Topic(topic).Publish(ctx, msg)

	// publishResult.Get() will make API call synchronous by awaiting messageId or error.
	// By removing .Get() method call we can make publishing asynchronous, but without ability to catch errors
	if _, err := publishResult.Get(context.Background()); err != nil {
		return errors.Wrap(err, "event publish failed")
	}

	return nil
}
