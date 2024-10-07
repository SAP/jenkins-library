package gcp

import (
	"cloud.google.com/go/pubsub"
	"context"
	piperConfig "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type PubsubClient interface {
	Publish(topic string, data []byte) error
}

type pubsubClient struct {
	vaultClient   piperConfig.VaultClient
	projectNumber string
	pool          string
	provider      string
	orderingKey   string
	oidcRoleId    string
}

func NewGcpPubsubClient(vaultClient piperConfig.VaultClient, projectNumber, pool, provider, orderingKey, oidcRoleId string) PubsubClient {
	return &pubsubClient{
		vaultClient:   vaultClient,
		projectNumber: projectNumber,
		pool:          pool,
		provider:      provider,
		orderingKey:   orderingKey,
		oidcRoleId:    oidcRoleId,
	}
}

func (cl *pubsubClient) Publish(topic string, data []byte) error {
	oidcToken, err := cl.vaultClient.GetOIDCTokenByValidation(cl.oidcRoleId)
	if err != nil {
		return errors.Wrap(err, "could not get oidc token")
	}

	accessToken, err := getFederatedToken(cl.projectNumber, cl.pool, cl.provider, oidcToken)
	if err != nil {
		return errors.Wrap(err, "could not get federated token")
	}

	return publish(cl.projectNumber, accessToken, topic, cl.orderingKey, data)
}

func publish(projectNumber, accessToken, topic, orderingKey string, data []byte) error {
	ctx := context.Background()

	staticTokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	pubsubClient, err := pubsub.NewClient(ctx, projectNumber, option.WithTokenSource(staticTokenSource))
	if err != nil {
		return errors.Wrap(err, "pubsub client creation failed")
	}

	t := pubsubClient.Topic(topic)
	t.EnableMessageOrdering = true
	publishResult := t.Publish(ctx, &pubsub.Message{Data: data, OrderingKey: orderingKey})

	// publishResult.Get() will make API call synchronous by awaiting messageId or error.
	// By removing .Get() method call we can make publishing asynchronous, but without ability to catch errors
	msgID, err := publishResult.Get(context.Background())
	if err != nil {
		return errors.Wrap(err, "event publish failed")
	}
	log.Entry().Debugf("Event published with ID: %s", msgID)

	return nil
}
