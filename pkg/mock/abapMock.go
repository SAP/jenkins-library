package mock

import "github.com/SAP/jenkins-library/pkg/abaputils"

// AUtilsMock mock
type AUtilsMock struct {
	ReturnedConnectionDetailsHTTP abaputils.ConnectionDetailsHTTP
	ReturnedError                 error
}

// GetAbapCommunicationArrangementInfo mock
func (abaputils *AUtilsMock) GetAbapCommunicationArrangementInfo(options abaputils.AbapEnvironmentOptions, oDataURL string) (abaputils.ConnectionDetailsHTTP, error) {
	return abaputils.ReturnedConnectionDetailsHTTP, abaputils.ReturnedError
}
