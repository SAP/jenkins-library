package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitCreateTargetVector(config abapAddonAssemblyKitCreateTargetVectorOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}
	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitCreateTargetVector(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCreateTargetVector(config *abapAddonAssemblyKitCreateTargetVectorOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment) error {
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	// var repos []abaputils.Repository
	// json.Unmarshal([]byte(config.Repositories), &repos)
	// var product abaputils.AddonDescriptor
	// json.Unmarshal([]byte(config.AddonProduct), &product)
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	var tv targetVector
	tv.init(addonDescriptor)

	err := tv.createTargetVector(*conn)
	if err != nil {
		return err
	}
	fmt.Println("after creation")
	fmt.Println(tv.ID)
	// TODO id zurück ins CPE
	addonDescriptor.TargetVectorID = tv.ID
	toCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(toCPE)
	return nil
}

func (tv *targetVector) createTargetVector(conn connector) error {
	conn.getToken("/odata/aas_ocs_package")
	tvJson, err := json.Marshal(tv)
	if err != nil {
		return err
	}
	appendum := "/odata/aas_ocs_package/TargetVectorSet"
	body, err := conn.post(appendum, string(tvJson))
	if err != nil {
		return err
	}
	var jTV targetVector
	json.Unmarshal(body, &jTV)
	fmt.Println(jTV.ID)
	tv.ID = jTV.ID
	return nil
}

func (tv *targetVector) init(addonDescriptor abaputils.AddonDescriptor) {
	tv.ProductName = addonDescriptor.AddonProduct
	tv.ProductVersion = addonDescriptor.AddonVersion
	tv.SpsLevel = addonDescriptor.AddonSpsLevel
	tv.PatchLevel = addonDescriptor.AddonPatchLevel

	var tvCVs []targetVectorCV
	var tvCV targetVectorCV
	for _, repo := range addonDescriptor.Repositories {
		tvCV.ScName = repo.Name
		tvCV.ScVersion = repo.Version
		tvCV.DeliveryPackage = repo.PackageName
		tvCV.SpLevel = repo.SpLevel
		tvCV.PatchLevel = repo.PatchLevel
		tvCVs = append(tvCVs, tvCV)
	}
	tv.Content.TargetVectorCVs = tvCVs
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
