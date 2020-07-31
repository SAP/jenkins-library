package abapMock

import "github.com/SAP/jenkins-library/pkg/abaputils"

// AUtilsMock mock
type AUtilsMock struct {
	ReturnedConnectionDetailsHTTP abaputils.ConnectionDetailsHTTP
	ReturnedError                 error
}

// GetAbapCommunicationArrangementInfo mock
func (autils *AUtilsMock) GetAbapCommunicationArrangementInfo(options abaputils.AbapEnvironmentOptions, oDataURL string) (abaputils.ConnectionDetailsHTTP, error) {
	return autils.ReturnedConnectionDetailsHTTP, autils.ReturnedError
}

func (autils *AUtilsMock) cleanup() {
	autils.ReturnedConnectionDetailsHTTP = abaputils.ConnectionDetailsHTTP{}
	autils.ReturnedError = nil
}
