package cmd

import (
	"encoding/json"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitCreateTargetVector(config abapAddonAssemblyKitCreateTargetVectorOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitCreateTargetVector(&config, telemetryData, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCreateTargetVector(config *abapAddonAssemblyKitCreateTargetVectorOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender, cpe *abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment) error {
	conn := new(abapbuild.Connector)
	conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	var tv targetVector
	err := tv.init(addonDescriptor)
	if err != nil {
		return err
	}
	log.Entry().Infof("Create target vector for product %s version %s", addonDescriptor.AddonProduct, addonDescriptor.AddonVersionYAML)
	err = tv.createTargetVector(*conn)
	if err != nil {
		return err
	}

	log.Entry().Infof("Created target vector %s", tv.ID)
	addonDescriptor.TargetVectorID = tv.ID
	log.Entry().Info("Write target vector to CommonPipelineEnvironment")
	toCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(toCPE)
	return nil
}

func (tv *targetVector) init(addonDescriptor abaputils.AddonDescriptor) error {
	if addonDescriptor.AddonProduct == "" || addonDescriptor.AddonVersion == "" || addonDescriptor.AddonSpsLevel == "" || addonDescriptor.AddonPatchLevel == "" {
		return errors.New("Parameters missing. Please provide product name, version, spslevel and patchlevel")
	}

	tv.ProductName = addonDescriptor.AddonProduct
	tv.ProductVersion = addonDescriptor.AddonVersion
	tv.SpsLevel = addonDescriptor.AddonSpsLevel
	tv.PatchLevel = addonDescriptor.AddonPatchLevel

	var tvCVs []targetVectorCV
	var tvCV targetVectorCV
	for i := range addonDescriptor.Repositories {
		if addonDescriptor.Repositories[i].Name == "" || addonDescriptor.Repositories[i].Version == "" || addonDescriptor.Repositories[i].SpLevel == "" ||
			addonDescriptor.Repositories[i].PatchLevel == "" || addonDescriptor.Repositories[i].PackageName == "" {
			return errors.New("Parameters missing. Please provide software component name, version, splevel, patchlevel and packagename")
		}
		tvCV.ScName = addonDescriptor.Repositories[i].Name
		tvCV.ScVersion = addonDescriptor.Repositories[i].Version
		tvCV.DeliveryPackage = addonDescriptor.Repositories[i].PackageName
		tvCV.SpLevel = addonDescriptor.Repositories[i].SpLevel
		tvCV.PatchLevel = addonDescriptor.Repositories[i].PatchLevel
		tvCVs = append(tvCVs, tvCV)
	}
	tv.Content.TargetVectorCVs = tvCVs
	return nil
}

func (tv *targetVector) createTargetVector(conn abapbuild.Connector) error {
	conn.GetToken("/odata/aas_ocs_package")
	tvJSON, err := json.Marshal(tv)
	if err != nil {
		return err
	}
	appendum := "/odata/aas_ocs_package/TargetVectorSet"
	body, err := conn.Post(appendum, string(tvJSON))
	if err != nil {
		return err
	}
	var jTV jsonTargetVector
	json.Unmarshal(body, &jTV)
	tv.ID = jTV.Tv.ID
	return nil
}

type jsonTargetVector struct {
	Tv *targetVector `json:"d"`
}

type targetVector struct {
	ID             string          `json:"Id"`
	ProductName    string          `json:"ProductName"`
	ProductVersion string          `json:"ProductVersion"`
	SpsLevel       string          `json:"SpsLevel"`
	PatchLevel     string          `json:"PatchLevel"`
	Content        targetVectorCVs `json:"Content"`
}

type targetVectorCV struct {
	ID              string `json:"Id"`
	ScName          string `json:"ScName"`
	ScVersion       string `json:"ScVersion"`
	DeliveryPackage string `json:"DeliveryPackage"`
	SpLevel         string `json:"SpLevel"`
	PatchLevel      string `json:"PatchLevel"`
}

type targetVectorCVs struct {
	TargetVectorCVs []targetVectorCV `json:"results"`
}
