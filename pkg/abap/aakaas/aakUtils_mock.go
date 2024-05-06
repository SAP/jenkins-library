//go:build !release
// +build !release

package aakaas

import (
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

type AakBundleMock struct {
	*mock.ExecMockRunner
	*abaputils.ClientMock
	*mock.FilesMock
	maxRuntime          time.Duration
	MockAddonDescriptor abaputils.AddonDescriptor
}

func NewAakBundleMock() *AakBundleMock {
	utils := AakBundleMock{
		ExecMockRunner: &mock.ExecMockRunner{},
		ClientMock:     &abaputils.ClientMock{},
		maxRuntime:     1 * time.Second,
		FilesMock:      &mock.FilesMock{},
	}
	return &utils
}

func (bundle *AakBundleMock) GetUtils() AakUtils {
	return bundle
}

func (bundle *AakBundleMock) GetMaxRuntime() time.Duration {
	return bundle.maxRuntime
}

func (bundle *AakBundleMock) SetMaxRuntime(maxRuntime time.Duration) {
	bundle.maxRuntime = maxRuntime
}

func (bundle *AakBundleMock) GetPollingInterval() time.Duration {
	return 1 * time.Microsecond
}

func (bundle *AakBundleMock) SetBodyList(bodyList []string) {
	bundle.ClientMock.Body = ""
	bundle.ClientMock.BodyList = bodyList
}

func (bundle *AakBundleMock) SetBody(body string) {
	bundle.ClientMock.Body = body
	bundle.ClientMock.BodyList = []string{}
}

func (bundle *AakBundleMock) SetErrorInsteadOfDumpToTrue() {
	bundle.ClientMock.ErrorInsteadOfDump = true
}

func (bundle *AakBundleMock) SetError(errorText string) {
	bundle.ClientMock.Error = errors.New(errorText)
}

func (bundle *AakBundleMock) ReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error) {
	var addonDescriptor abaputils.AddonDescriptor
	var err error
	switch FileName {
	case "success":
		{
			addonDescriptor = abaputils.AddonDescriptor{
				AddonProduct:     "/DRNMSPC/PRD01",
				AddonVersionYAML: "3.2.1",
				Repositories: []abaputils.Repository{
					{
						Name:        "/DRNMSPC/COMP01",
						VersionYAML: "1.2.3",
						CommitID:    "HUGO1234",
					},
				},
			}
		}
	case "noCommitID":
		{
			addonDescriptor = abaputils.AddonDescriptor{
				AddonProduct:     "/DRNMSPC/PRD01",
				AddonVersionYAML: "3.2.1",
				Repositories: []abaputils.Repository{
					{
						Name:        "/DRNMSPC/COMP01",
						VersionYAML: "1.2.3",
					},
				},
			}
		}
	case "failing":
		{
			err = errors.New("error in ReadAddonDescriptor")
		}
	case "addon.yml.mock":
		{
			return bundle.MockAddonDescriptor, nil
		}
	}
	return addonDescriptor, err
}

// *****************************other client mock *******************************
type AakBundleMockNewMC struct {
	*mock.ExecMockRunner
	*abapbuild.MockClient
	*piperutils.Files
	maxRuntime time.Duration
}

func NewAakBundleMockNewMC(mC *abapbuild.MockClient) *AakBundleMockNewMC {
	utils := AakBundleMockNewMC{
		ExecMockRunner: &mock.ExecMockRunner{},
		MockClient:     mC,
		maxRuntime:     1 * time.Second,
	}
	return &utils
}

func (bundle *AakBundleMockNewMC) GetUtils() AakUtils {
	return bundle
}

func (bundle *AakBundleMockNewMC) GetMaxRuntime() time.Duration {
	return bundle.maxRuntime
}

func (bundle *AakBundleMockNewMC) GetPollingInterval() time.Duration {
	return 1 * time.Microsecond
}

func (bundle *AakBundleMockNewMC) ReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error) {
	var addonDescriptor abaputils.AddonDescriptor
	err := errors.New("don't use this")
	return addonDescriptor, err
}
