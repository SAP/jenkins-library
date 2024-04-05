package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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

	if err := pvh.checkAndResolveVersion(conn); err != nil {
		return err
	}
	pvh.SyncAddonDescriptorVersionFields(&addonDescriptor)

	commonPipelineEnvironment.abap.addonDescriptor = string(addonDescriptor.AsJSON())

	publishAddonYaml(config, utils)

	return nil
}

func publishAddonYaml(config *abapAddonAssemblyKitCheckOptions, utils aakaas.AakUtils) {
	var filesToPublish []piperutils.Path
	log.Entry().Infof("Add %s to be published", config.AddonDescriptorFileName)
	filesToPublish = append(filesToPublish, piperutils.Path{Target: config.AddonDescriptorFileName, Name: "AddonDescriptor", Mandatory: true})
	log.Entry().Infof("Publishing %v files", len(filesToPublish))
	if err := piperutils.PersistReportsAndLinks("abapAddonAssemblyKitCheckPV", "", utils, filesToPublish, nil); err != nil {
		log.Entry().WithError(err).Error("failed to persist report information")
	}
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

func (pv *ProductVersionHeader) checkAndResolveVersion(conn *abapbuild.Connector) error {
	requestJson, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	appendum := "/odata/aas_ocs_package/ProductVersionHeaderSet"
	responseBody, err := conn.Post(appendum, string(requestJson))
	if err != nil {
		return errors.Wrap(err, "Checking Product Modeling in AAkaaS failed")
	}
	var resultPv ProductVersionHeader //wrapper struct needed???
	if err := json.Unmarshal(responseBody, &resultPv); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for checking Product Modeling "+string(responseBody))
	}

	pv.ProductVersion = resultPv.ProductVersion
	pv.Spslevel = resultPv.Spslevel
	pv.PatchLevel = resultPv.PatchLevel

	//transfer resolved version fields (and others)
	for pvc_index, pvc := range pv.Content {
		foundPvc := ProductVersionContent{}
		for _, resultPvc := range resultPv.Content {
			if pvc.SoftwareComponentName == resultPvc.SoftwareComponentName && foundPvc.SoftwareComponentName == "" {
				foundPvc = resultPvc
			} else if pvc.SoftwareComponentName == resultPvc.SoftwareComponentName {
				return errors.New("Software Component Name must be unique in the ProductVersionContent")
			}
		}
		if foundPvc.SoftwareComponentName == "" {
			return errors.New("Software Component Name not found in the ProductVersionContent")
		}
		pv.Content[pvc_index].PatchLevel = foundPvc.PatchLevel
		pv.Content[pvc_index].SpLevel = foundPvc.SpLevel
		pv.Content[pvc_index].SoftwareComponentVersion = foundPvc.SoftwareComponentVersion
	}

	return nil
}

func (pv *ProductVersionHeader) SyncAddonDescriptorVersionFields(addonDescriptor *abaputils.AddonDescriptor) error {
	addonDescriptor.AddonVersion = pv.ProductVersion
	addonDescriptor.AddonSpsLevel = pv.Spslevel
	addonDescriptor.AddonPatchLevel = pv.PatchLevel

	//in NewPvh function pvh was build up 1:1 based on addonDescriptor
	//in checkAndResolve pvh was synced from AAKaaS reply assuming it does not contain more content than before(if it does it is ignored)
	//anyway we are defenisve once more
	for repo_index, repo := range addonDescriptor.Repositories {
		foundPvc := ProductVersionContent{}
		for _, pvc := range pv.Content {
			if pvc.SoftwareComponentName == repo.Name && foundPvc.SoftwareComponentName == "" {
				foundPvc = pvc
			} else if pvc.SoftwareComponentName == repo.Name {
				return errors.New("Software Component Name must be unique in addon descriptor(aka addon.yml)")
			}
		}
		if foundPvc.SoftwareComponentName == "" {
			return errors.New("ProductVersionContent & addon descriptor (aka addon.yml) out of sync")
		}

		addonDescriptor.Repositories[repo_index].PatchLevel = foundPvc.PatchLevel
		addonDescriptor.Repositories[repo_index].SpLevel = foundPvc.SpLevel
		addonDescriptor.Repositories[repo_index].Version = foundPvc.SoftwareComponentVersion
	}

	return nil
}
