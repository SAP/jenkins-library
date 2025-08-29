package feature

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

const prefix = "com_sap_piper_featureFlag_"

func IsFeatureEnabled(flag string) bool {
	if os.Getenv(prefix+flag) == "true" {
		log.Entry().Infof("Feature '%s%s' is enabled", prefix, flag)
		return true
	}
	return false
}
