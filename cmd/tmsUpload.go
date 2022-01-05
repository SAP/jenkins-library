package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/tms"
)

type uaa struct {
	Url          string `json:"url"`
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
}

type serviceKey struct {
	Uaa uaa    `json:"uaa"`
	Uri string `json:"uri"`
}

type tmsUploadUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The tmsUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type tmsUploadUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to tmsUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// tmsUploadUtilsBundle and forward to the implementation of the dependency.
}

func newTmsUploadUtils() tmsUploadUtils {
	utils := tmsUploadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func tmsUpload(config tmsUploadOptions, telemetryData *telemetry.CustomData, influx *tmsUploadInflux) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	// utils := newTmsUploadUtils()

	proxy := config.Proxy
	transportProxy, err := url.Parse(proxy)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to parse proxy string %v into a URL structure", proxy)
	}
	options := piperHttp.ClientOptions{TransportProxy: transportProxy}

	client := &piperHttp.Client{}
	client.SetOptions(options)
	if GeneralConfig.Verbose && proxy != "" {
		log.Entry().Infof("HTTP client instructed to use %v proxy", proxy)
	}

	serviceKey, err := unmarshalServiceKey(config.TmsServiceKey)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to unmarshal TMS service key")
	}

	communicationInstance, err := tms.NewCommunicationInstance(client, serviceKey.Uri, serviceKey.Uaa.Url, serviceKey.Uaa.ClientId, serviceKey.Uaa.ClientSecret, GeneralConfig.Verbose)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to prepare client for talking with TMS")
	}

	// TODO: understand, what does this influx part do
	influx.step_data.fields.tms = false

	if err := runTmsUpload(config, communicationInstance); err != nil {
		log.Entry().WithError(err).Fatal("Failed to run tmsUpload step")
	}
	influx.step_data.fields.tms = true
}

func runTmsUpload(config tmsUploadOptions, communicationInstance tms.CommunicationInterface) error {
	// TODO: provide TMS upload logic here
	/*
		log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

		// Example of calling methods from external dependencies directly on utils:
		exists, err := utils.FileExists("file.txt")
		if err != nil {
			// It is good practice to set an error category.
			// Most likely you want to do this at the place where enough context is known.
			log.SetErrorCategory(log.ErrorConfiguration)
			// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
			return fmt.Errorf("failed to check for important file: %w", err)
		}
		if !exists {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("cannot run without important file")
		}
	*/

	// TODO: this is very simple and hardcoded for now
	// 1. Get nodes
	nodes, err1 := communicationInstance.GetNodes()
	if err1 != nil {
		return fmt.Errorf("failed to get nodes: %w", err1)
	}

	var nodeId int64
	for i := 0; i < len(nodes); i++ {
		if nodes[i].Name == config.NodeName {
			nodeId = nodes[i].Id
			break
		}
	}

	// 2. Get MTA extension descriptor for node with given id
	obtainedMtaExtDescriptor, err2 := communicationInstance.GetMtaExtDescriptor(nodeId, "com.sap.radsoulbeard.fiori", config.MtaVersion)
	if err2 != nil {
		return fmt.Errorf("failed to get MTA extension descriptor: %w", err2)
	}

	// 3. Update obtained MTA extension descriptor
	// TODO: read MTA extension descriptor file path from the map provided in config.yml
	_, err3 := communicationInstance.UpdateMtaExtDescriptor(nodeId, obtainedMtaExtDescriptor.Id, "mtaext_update_test.mtaext", config.MtaVersion, config.CustomDescription, config.NamedUser)
	if err3 != nil {
		return fmt.Errorf("failed to update MTA extension descriptor: %w", err3)
	}

	// 4. Upload another MTA extension descriptor to node
	_, err4 := communicationInstance.UploadMtaExtDescriptorToNode(nodeId, "mtaext_upload_test.mtaext", "1.0.1", config.CustomDescription, config.NamedUser)
	if err4 != nil {
		return fmt.Errorf("failed to upload MTA extension descriptor: %w", err4)
	}

	// 5. Upload file
	fileInfo, err5 := communicationInstance.UploadFile(config.MtaPath, config.NamedUser)
	if err5 != nil {
		return fmt.Errorf("failed to upload file: %w", err5)
	}

	// 6. Uplaod file to node
	_, err6 := communicationInstance.UploadFileToNode(config.NodeName, strconv.FormatInt(fileInfo.Id, 10), config.CustomDescription, config.NamedUser)
	if err6 != nil {
		return fmt.Errorf("failed to upload file to node: %w", err6)
	}

	return nil
}

func unmarshalServiceKey(serviceKeyJson string) (serviceKey serviceKey, err error) {
	err = json.Unmarshal([]byte(serviceKeyJson), &serviceKey)
	if err != nil {
		return
	}
	return
}
