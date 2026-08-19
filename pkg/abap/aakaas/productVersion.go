package aakaas

import (
	"encoding/json"
	"fmt"
	"net/url"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
)

const pvQueryURL string = "/odata/aas_ocs_package/xSSDAxC_Product_Version"
const pvValidateURL string = "/odata/aas_ocs_package/ValidateProductVersion"

type ProductVersion struct {
	versionable
}

func (p *ProductVersion) ConstructProductversion(desc abaputils.AddonDescriptor, conn abapbuild.Connector) error {
	if err := p.constructVersionable(desc.AddonProduct, desc.AddonVersionYAML, conn, pvQueryURL); err != nil {
		return err
	}
	if err := p.resolveWildCards(statusFilterPV); err != nil {
		return err
	}
	return nil
}

func (p *ProductVersion) CopyVersionFieldsToDescriptor(desc *abaputils.AddonDescriptor) {
	desc.AddonVersion = p.TechRelease
	desc.AddonSpsLevel = p.TechSpLevel
	desc.AddonPatchLevel = p.TechPatchLevel
	desc.AddonVersionYAML = p.Version
}

func (p *ProductVersion) ValidateAndResolveVersionFields() error {
	log.Entry().Infof("Validate product '%s' version '%s' and resolve version", p.Name, p.Version)

	values := url.Values{}
	values.Set("Name", "'"+p.Name+"'")
	values.Set("Version", "'"+p.Version+"'")
	requestUrl := pvValidateURL + "?" + values.Encode()

	body, err := p.connector.Get(requestUrl)
	if err != nil {
		return err
	}
	var response jsonProductVersionValidationResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("Unexpected AAKaaS response for Validate Product Version: "+string(body), err)
	}
	p.Name = response.Wrapper.Name
	p.TechRelease = response.Wrapper.TechRelease
	p.TechSpLevel = response.Wrapper.TechSpLevel
	p.TechPatchLevel = response.Wrapper.TechPatchLevel
	log.Entry().Infof("Resolved version %s, spslevel %s, patchlevel %s", p.TechRelease, p.TechSpLevel, p.TechPatchLevel)
	return nil
}

type jsonProductVersionValidationResponse struct {
	Wrapper struct {
		Name           string `json:"Name"`
		TechRelease    string `json:"Version"`
		TechSpLevel    string `json:"SpsLevel"`
		TechPatchLevel string `json:"PatchLevel"`
	} `json:"d"`
}
