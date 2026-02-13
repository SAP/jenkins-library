package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"errors"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitRegisterPackages(config abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}
	telemetryData.BuildTool = "AAKaaS"

	if err := runAbapAddonAssemblyKitRegisterPackages(&config, &client, cpe, reader); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitRegisterPackages(config *abapAddonAssemblyKitRegisterPackagesOptions, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment, fileReader readFile) error {

	var addonDescriptor abaputils.AddonDescriptor
	if err := json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor); err != nil {
		return err
	}

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client, config.AbapAddonAssemblyKitOriginHash, config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	if err := uploadSarFiles(addonDescriptor.Repositories, *conn, fileReader); err != nil {
		return err
	}

	conn2 := new(abapbuild.Connector) // we need a second connector without the added Header
	if err := conn2.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client, config.AbapAddonAssemblyKitOriginHash, config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	var err error
	addonDescriptor.Repositories, err = registerPackages(addonDescriptor.Repositories, *conn2)
	if err != nil {
		return err
	}

	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	cpe.abap.addonDescriptor = addonDescriptor.AsJSONstring()

	return nil
}

func uploadSarFiles(repos []abaputils.Repository, conn abapbuild.Connector, readFileFunc readFile) error {
	for i := range repos {
		if repos[i].Status == string(aakaas.PackageStatusPlanned) {
			if repos[i].SarXMLFilePath == "" {
				return errors.New("Parameter missing. Please provide the path to the SAR file")
			}
			filename := filepath.Base(repos[i].SarXMLFilePath)
			log.Entry().Infof("Trying to read file %s", repos[i].SarXMLFilePath)
			sarFile, err := readFileFunc(repos[i].SarXMLFilePath)
			if err != nil {
				return err
			}
			log.Entry().Infof("... %d bytes read", len(sarFile))
			if len(sarFile) == 0 {
				return errors.New("File has no content - 0 bytes")
			}
			log.Entry().Infof("Upload SAR file %s in chunks", filename)
			err = conn.UploadSarFileInChunks("/odata/aas_file_upload", filename, sarFile)
			if err != nil {
				return err
			}
			log.Entry().Info("...done")
		} else {
			log.Entry().Infof("Package %s has status %s, cannot upload the SAR file of this package", repos[i].PackageName, repos[i].Status)
		}
	}
	return nil
}

// for moocking
type readFile func(path string) ([]byte, error)

func reader(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func registerPackages(repos []abaputils.Repository, conn abapbuild.Connector) ([]abaputils.Repository, error) {
	for i := range repos {
		var pack aakaas.Package
		pack.InitPackage(repos[i], conn)
		if repos[i].Status == string(aakaas.PackageStatusPlanned) {
			log.Entry().Infof("Trying to Register Package %s", pack.PackageName)
			err := pack.Register()
			if err != nil {
				return repos, err
			}
			log.Entry().Info("...done, take over new status")
			pack.ChangeStatus(&repos[i])
			log.Entry().Info("...done")
		} else {
			log.Entry().Infof("Package %s has status %s, cannot register this package", pack.PackageName, pack.Status)
		}
	}
	return repos, nil
}
