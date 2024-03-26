package cmd

import (
	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitCheck(config abapAddonAssemblyKitCheckOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *abapAddonAssemblyKitCheckCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundle()

	err := runAbapAddonAssemblyKitCheck(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheck(config *abapAddonAssemblyKitCheckOptions, telemetryData *telemetry.CustomData, utils aakaas.AakUtils, commonPipelineEnvironment *abapAddonAssemblyKitCheckCommonPipelineEnvironment) error {

	log.Entry().Info("╔═══════════════════════════╗")
	log.Entry().Info("║ abapAddonAssemblyKitCheck ║")
	log.Entry().Info("╚═══════════════════════════╝")

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, utils, "", config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	log.Entry().Infof("Reading addonDescriptor (aka addon.yml) file: %s", config.AddonDescriptorFileName)
	addonDescriptor, err := utils.ReadAddonDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return err
	}

	pvh, err := NewProductVersionHeader(&addonDescriptor, conn)
	if err != nil {
		return err
	}

	if err := pvh.check(); err != nil {
		return err
	}

	// log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	// // Example of calling methods from external dependencies directly on utils:
	// exists, err := utils.FileExists("file.txt")
	// if err != nil {
	// 	// It is good practice to set an error category.
	// 	// Most likely you want to do this at the place where enough context is known.
	// 	log.SetErrorCategory(log.ErrorConfiguration)
	// 	// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
	// 	return fmt.Errorf("failed to check for important file: %w", err)
	// }
	// if !exists {
	// 	log.SetErrorCategory(log.ErrorConfiguration)
	// 	return fmt.Errorf("cannot run without important file")
	// }

	return nil
}

type ProductVersionHeader struct {
	ProductName            string
	SemanticProductVersion string `json:"SemProductVersion"`
	ProductVersion         string
	Spslevel               string
	PatchLevel             string
	Vendor                 string
	VendorType             string
	Content                []ProductVersionContent //maybe some struct in between see TargetVector???
}

type ProductVersionContent struct {
	ProductName                      string
	SemanticProductVersion           string `json:"SemProductVersion"`
	SoftwareComponentName            string `json:"ScName"`
	SemanticSoftwareComponentVersion string `json:"SemScVersion"`
	SoftwareComponentVersion         string `json:"ScVersion"`
	SpLevel                          string
	PatchLevel                       string
	Vendor                           string
	VendorType                       string
}

func NewProductVersionHeader(addonDescriptor *abaputils.AddonDescriptor, conn *abapbuild.Connector) (*ProductVersionHeader, error) {
	productVersion := new(aakaas.ProductVersion)
	if err := productVersion.ConstructProductversion(*addonDescriptor, *conn); err != nil {
		return nil, err
	}
	pvh := ProductVersionHeader{
		ProductName:            productVersion.Name,
		SemanticProductVersion: productVersion.Version,
		Content:                []ProductVersionContent{},
	}

	for _, repo := range addonDescriptor.Repositories {
		componentVersion := new(aakaas.ComponentVersion)
		if err := componentVersion.ConstructComponentVersion(repo, *conn); err != nil {
			return nil, err
		}
		pvc := ProductVersionContent{
			ProductName:                      pvh.ProductName,
			SemanticProductVersion:           pvh.ProductName,
			SoftwareComponentName:            componentVersion.Name,
			SemanticSoftwareComponentVersion: componentVersion.Version,
		}
		pvh.Content = append(pvh.Content, pvc)
	}

	if len(pvh.Content) == 0 {
		return nil, errors.New("addonDescriptor must contain at least one software component repository")
	}

	return &pvh, nil
}

func (pv *ProductVersionHeader) check() error {

	return nil
}
