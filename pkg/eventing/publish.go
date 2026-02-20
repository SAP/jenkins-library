// Package eventing is a placeholder for eventing related code, e.g. publish cloud events to GCP Pub/Sub might be deleted
package eventing

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
)

func Publish(tokenProvider gcp.OIDCTokenProvider, GeneralConfig config.GeneralConfigOptions, eventData []byte) error {
	// publish cloud event via GCP Pub/Sub
	err := gcp.NewGcpPubsubClient(
		tokenProvider,
		GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
		GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
		GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
		GeneralConfig.CorrelationID,
		GeneralConfig.HookConfig.OIDCConfig.RoleID,
	).Publish(
		fmt.Sprintf("%spipelinetaskrun-finished", GeneralConfig.HookConfig.GCPPubSubConfig.TopicPrefix),
		eventData,
	)
	if err != nil {
		return err
	}
	return nil
}