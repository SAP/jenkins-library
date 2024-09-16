package aakaas

import (
	"encoding/json"
	"net/url"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

const cvQueryURL string = "/odata/aas_ocs_package/xSSDAxC_Component_Version"
const cvValidateURL string = "/odata/aas_ocs_package/ValidateComponentVersion"

type ComponentVersion struct {
	versionable
}

func (c *ComponentVersion) ConstructComponentVersion(repo abaputils.Repository, conn abapbuild.Connector) error {
	if err := c.constructVersionable(repo.Name, repo.VersionYAML, conn, cvQueryURL); err != nil {
		return err
	}
	if err := c.resolveWildCards(statusFilterCV); err != nil {
		return err
	}

	return nil
}

func (c *ComponentVersion) CopyVersionFieldsToRepo(repo *abaputils.Repository) {
	repo.Version = c.TechRelease
	repo.SpLevel = c.TechSpLevel
	repo.PatchLevel = c.TechPatchLevel
	repo.VersionYAML = c.Version
}

func (c *ComponentVersion) Validate() error {
	log.Entry().Infof("Validate component %s version %s and resolve version", c.Name, c.Version)

	values := url.Values{}
	values.Set("Name", "'"+c.Name+"'")
	values.Set("Version", "'"+c.Version+"'")
	requestUrl := cvValidateURL + "?" + values.Encode()

	body, err := c.connector.Get(requestUrl)
	if err != nil {
		return err
	}
	var response jsonComponentVersionValidationResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for Validate Component Version: "+string(body))
	}
	c.Name = response.Wrapper.Name
	c.TechRelease = response.Wrapper.TechRelease
	c.TechSpLevel = response.Wrapper.TechSpLevel
	c.TechPatchLevel = response.Wrapper.TechPatchLevel
	log.Entry().Infof("Resolved version %s, splevel %s, patchlevel %s", c.TechRelease, c.TechSpLevel, c.TechPatchLevel)

	return nil
}

type jsonComponentVersionValidationResponse struct {
	Wrapper struct {
		Name           string `json:"Name"`
		TechRelease    string `json:"Version"`
		TechSpLevel    string `json:"SpLevel"`
		TechPatchLevel string `json:"PatchLevel"`
	} `json:"d"`
}
