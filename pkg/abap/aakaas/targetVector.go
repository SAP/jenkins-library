package aakaas

import (
	"encoding/json"
	"net/url"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
)

// TargetVectorStatus : Status of TargetVector in AAKaaS
type TargetVectorStatus string

const (
	// TargetVectorStatusGenerated : TargetVector is Generated (not published yet)
	TargetVectorStatusGenerated TargetVectorStatus = "G"
	// TargetVectorStatusTest : TargetVector is published for testing
	TargetVectorStatusTest TargetVectorStatus = "T"
	// TargetVectorStatusProductive : TargetVector is published for productive use
	TargetVectorStatusProductive TargetVectorStatus = "P"

	TargetVectorPublishStatusRunning TargetVectorStatus = "R"
	TargetVectorPublishStatusSuccess TargetVectorStatus = "S"
	TargetVectorPublishStatusError   TargetVectorStatus = "E"
)

// internal usage : Json Structure for AAKaaS Odata Service
type jsonTargetVector struct {
	Tv *TargetVector `json:"d"`
}

// TargetVector : TargetVector desribes a deployble state of an ABAP product version
type TargetVector struct {
	ID             string             `json:"Id"`
	ProductName    string             `json:"ProductName"`
	ProductVersion string             `json:"ProductVersion"`
	SpsLevel       string             `json:"SpsLevel"`
	PatchLevel     string             `json:"PatchLevel"`
	Status         TargetVectorStatus `json:"Status"`
	PublishStatus  TargetVectorStatus `json:"PublishStatus"`
	Content        TargetVectorCVs    `json:"Content"`
}

// TargetVectorCV : deployable state of an software Component Version as part of an TargetVector
type TargetVectorCV struct {
	ID              string `json:"Id"`
	ScName          string `json:"ScName"`
	ScVersion       string `json:"ScVersion"`
	DeliveryPackage string `json:"DeliveryPackage"`
	SpLevel         string `json:"SpLevel"`
	PatchLevel      string `json:"PatchLevel"`
}

// TargetVectorCVs : deployable states of the software Component Versions of the product version
type TargetVectorCVs struct {
	TargetVectorCVs []TargetVectorCV `json:"results"`
}

// Init : Initialize TargetVector for Creation in AAKaaS
func (tv *TargetVector) InitNew(addonDescriptor *abaputils.AddonDescriptor) error {
	if addonDescriptor.AddonProduct == "" || addonDescriptor.AddonVersion == "" || addonDescriptor.AddonSpsLevel == "" || addonDescriptor.AddonPatchLevel == "" {
		return errors.New("Parameters missing. Please provide product name, version, spslevel and patchlevel")
	}
	tv.ProductName = addonDescriptor.AddonProduct
	tv.ProductVersion = addonDescriptor.AddonVersion
	tv.SpsLevel = addonDescriptor.AddonSpsLevel
	tv.PatchLevel = addonDescriptor.AddonPatchLevel

	var tvCVs []TargetVectorCV
	var tvCV TargetVectorCV
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

// InitExisting : Initialize an already in AAKaaS existing TargetVector
func (tv *TargetVector) InitExisting(ID string) {
	tv.ID = ID
}

// CreateTargetVector : Initial Creation of an TargetVector
func (tv *TargetVector) CreateTargetVector(conn *abapbuild.Connector) error {
	conn.GetToken("/odata/aas_ocs_package")
	tvJSON, err := json.Marshal(tv)
	if err != nil {
		return errors.Wrap(err, "Generating Request Data for Create Target Vector failed")
	}
	appendum := "/odata/aas_ocs_package/TargetVectorSet"
	body, err := conn.Post(appendum, string(tvJSON))
	if err != nil {
		return errors.Wrap(err, "Creating Target Vector in AAKaaS failed")
	}
	var jTV jsonTargetVector
	if err := json.Unmarshal(body, &jTV); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for create target vector: "+string(body))
	}
	tv.ID = jTV.Tv.ID
	tv.Status = jTV.Tv.Status
	return nil
}

func (tv *TargetVector) PublishTargetVector(conn *abapbuild.Connector, targetVectorScope TargetVectorStatus) error {
	conn.GetToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/PublishTargetVector?Id='" + url.QueryEscape(tv.ID) + "'&Scope='" + url.QueryEscape(string(targetVectorScope)) + "'"
	body, err := conn.Post(appendum, "")
	if err != nil {
		return errors.Wrap(err, "Publish Target Vector in AAKaaS failed")
	}

	var jTV jsonTargetVector
	if err := json.Unmarshal(body, &jTV); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for publish target vector: "+string(body))
	}

	tv.Status = jTV.Tv.Status
	tv.PublishStatus = jTV.Tv.PublishStatus
	return nil
}

// GetTargetVector : Read details of the TargetVector
func (tv *TargetVector) GetTargetVector(conn *abapbuild.Connector) error {
	if tv.ID == "" {
		return errors.New("Without ID no details of a targetVector can be obtained from AAKaaS")
	}
	appendum := "/odata/aas_ocs_package/TargetVectorSet('" + url.QueryEscape(tv.ID) + "')"
	body, err := conn.Get(appendum)
	if err != nil {
		return errors.Wrap(err, "Getting Target Vector details from AAKaaS failed")
	}

	var jTV jsonTargetVector
	if err := json.Unmarshal(body, &jTV); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for getting target vector details: "+string(body))
	}

	tv.Status = jTV.Tv.Status
	tv.PublishStatus = jTV.Tv.PublishStatus
	//other fields not needed atm
	return nil
}

// PollForStatus : Poll AAKaaS until final PublishStatus reached and check if desired Status was reached
func (tv *TargetVector) PollForStatus(conn *abapbuild.Connector, targetStatus TargetVectorStatus) error {
	timeout := time.After(conn.MaxRuntimeInMinutes)
	ticker := time.Tick(conn.PollIntervalsInSeconds)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out (AAKaaS target Vector Status change)")
		case <-ticker:
			tv.GetTargetVector(conn)
			switch TargetVectorStatus(tv.PublishStatus) {
			case TargetVectorPublishStatusRunning:
				continue
			case TargetVectorPublishStatusSuccess:
				if TargetVectorStatus(tv.Status) == targetStatus {
					return nil
				} else {
					return errors.New("Publishing of Targetvector " + tv.ID + " resulted in state " + string(tv.Status) + "instead of expected state " + string(targetStatus))
				}
			case TargetVectorPublishStatusError:
				return errors.New("Publishing of Targetvector " + tv.ID + " failed in AAKaaS")
			}
		}
	}
}
