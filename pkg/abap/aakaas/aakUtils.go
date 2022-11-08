package aakaas

import (
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type AakUtils interface {
	command.ExecRunner
	abapbuild.HTTPSendLoader
	piperutils.FileUtils
	ReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error)
	GetMaxRuntime() time.Duration
	GetPollingInterval() time.Duration
}

type AakBundle struct {
	*command.Command
	*piperhttp.Client
	*piperutils.Files
	maxRuntime      time.Duration
	pollingInterval time.Duration
}

func (bundle *AakBundle) GetMaxRuntime() time.Duration {
	return bundle.maxRuntime
}

func (bundle *AakBundle) GetPollingInterval() time.Duration {
	return bundle.pollingInterval
}

func (bundle *AakBundle) ReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error) {
	return abaputils.ReadAddonDescriptor(FileName)
}

func NewAakBundleWithTime(maxRuntime time.Duration, pollingInterval time.Duration) AakUtils {
	utils := AakBundle{
		Command:         &command.Command{},
		Client:          &piperhttp.Client{},
		maxRuntime:      maxRuntime * time.Minute,
		pollingInterval: pollingInterval * time.Second,
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func NewAakBundle() AakUtils {
	utils := AakBundle{
		Command: &command.Command{},
		Client:  &piperhttp.Client{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}
