package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"github.com/SAP/jenkins-library/pkg/log"
	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type OIDCTokenProvider func(roleID string) (string, error)

type PubsubClient interface {
	// Publish encodes the CloudEvent into a Pub/Sub message via the CloudEvents
	// SDK and sends it to the topic. Callers pass the event directly; no manual
	// JSON marshaling is required.
	Publish(topic string, event cloudevents.Event) error
}

type pubsubClient struct {
	tokenProvider OIDCTokenProvider
	projectNumber string
	pool          string
	provider      string
	orderingKey   string
	oidcRoleId    string
}

func NewGcpPubsubClient(tokenProvider OIDCTokenProvider, projectNumber, pool, provider, orderingKey, oidcRoleId string) PubsubClient {
	return &pubsubClient{
		tokenProvider: tokenProvider,
		projectNumber: projectNumber,
		pool:          pool,
		provider:      provider,
		orderingKey:   orderingKey,
		oidcRoleId:    oidcRoleId,
	}
}

func (cl *pubsubClient) Publish(topic string, event cloudevents.Event) error {
	ctx := context.Background()
	psClient, err := cl.getAuthorizedGCPClient(ctx)
	if err != nil {
		return fmt.Errorf("could not get authorized pubsub client token: %w", err)
	}

	return cl.publish(ctx, psClient, topic, cl.orderingKey, event)
}

func (cl *pubsubClient) getAuthorizedGCPClient(ctx context.Context) (*pubsub.Client, error) {
	oidcToken, err := cl.tokenProvider(cl.oidcRoleId)
	if err != nil {
		return nil, fmt.Errorf("could not get oidc token: %w", err)
	}

	accessToken, err := getFederatedToken(cl.projectNumber, cl.pool, cl.provider, oidcToken)
	if err != nil {
		return nil, fmt.Errorf("could not get federated token: %w", err)
	}

	staticTokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	return pubsub.NewClient(ctx, cl.projectNumber, option.WithTokenSource(staticTokenSource))
}

func (cl *pubsubClient) publish(ctx context.Context, psClient *pubsub.Client, topic, orderingKey string, event cloudevents.Event) error {
	msg := &pubsub.Message{
		Attributes:  map[string]string{},
		OrderingKey: orderingKey,
	}
	if err := cepubsub.WritePubSubMessage(ctx, binding.ToMessage(&event), msg); err != nil {
		return fmt.Errorf("could not encode CloudEvent into pubsub message: %w", err)
	}

	t := psClient.Publisher(topic)
	t.EnableMessageOrdering = true
	publishResult := t.Publish(ctx, msg)

	// publishResult.Get() will make API call synchronous by awaiting messageId or error.
	// By removing .Get() method call we can make publishing asynchronous, but without ability to catch errors
	msgID, err := publishResult.Get(ctx)
	if err != nil {
		return fmt.Errorf("event publish failed: %w", err)
	}

	log.Entry().Debugf("Event published with ID: %s", msgID)
	return nil
}
