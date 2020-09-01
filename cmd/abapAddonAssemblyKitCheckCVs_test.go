package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestInitCV(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckCVs{}
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
		}
		var c cv
		c.init(repo, *conn)
		assert.Equal(t, "/DRNMSPC/COMP01", c.Name)
		assert.Equal(t, "1.2.3", c.VersionYAML)
	})
}

func TestValidateCV(t *testing.T) {
	t.Run("test validate", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckCVs{}
		var c cv
		c.connector = *conn
		c.Name = "/DRNMSPC/COMP01"
		c.VersionYAML = "1.2.3"
		err := c.validate()
		assert.NoError(t, err)
		assert.Equal(t, c.Version, "0001")
		assert.Equal(t, c.SpLevel, "0002")
		assert.Equal(t, c.PatchLevel, "0003")
	})
}

func TestValidateError(t *testing.T) {
	t.Run("test validate with error", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckCVs{}
		var c cv
		c.connector = *conn
		c.Name = "ERROR"
		c.VersionYAML = "1.2.3"
		err := c.validate()
		assert.Error(t, err)
		assert.Equal(t, c.Version, "")
		assert.Equal(t, c.SpLevel, "")
		assert.Equal(t, c.PatchLevel, "")
	})
}

func TestCopyFields(t *testing.T) {
	t.Run("test copyFieldsToRepo", func(t *testing.T) {
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
		}
		var c cv
		c.Version = "0001"
		c.SpLevel = "0002"
		c.PatchLevel = "0003"
		c.copyFieldsToRepo(&repo)
		assert.Equal(t, "0001", repo.Version)
		assert.Equal(t, "0002", repo.SpLevel)
		assert.Equal(t, "0003", repo.PatchLevel)
	})
}

// combineYAMLRepositoriesWithCPEProduct(addonDescriptor abaputils.AddonDescriptor, addonDescriptorFromCPE abaputils.AddonDescriptor) abaputils.AddonDescriptor
func TestCombineYAMLRepositoriesWithCPEProduct(t *testing.T) {
	t.Run("test combineYAMLRepositoriesWithCPEProduct", func(t *testing.T) {
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.2.3",
				},
				{
					Name:        "/DRNMSPC/COMP02",
					VersionYAML: "3.2.1",
				},
			},
		}
		addonDescriptorFromCPE := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSP/PROD",
			AddonVersionYAML: "1.2.3",
		}
		finalAddonDescriptor := combineYAMLRepositoriesWithCPEProduct(addonDescriptor, addonDescriptorFromCPE)
		assert.Equal(t, "/DRNMSP/PROD", finalAddonDescriptor.AddonProduct)
		assert.Equal(t, "1.2.3", finalAddonDescriptor.AddonVersionYAML)
		assert.Equal(t, "/DRNMSPC/COMP01", finalAddonDescriptor.Repositories[0].Name)
		assert.Equal(t, "/DRNMSPC/COMP02", finalAddonDescriptor.Repositories[1].Name)
		assert.Equal(t, "1.2.3", finalAddonDescriptor.Repositories[0].VersionYAML)
		assert.Equal(t, "3.2.1", finalAddonDescriptor.Repositories[1].VersionYAML)
	})
}

type clMockCheckCVs struct {
	StatusCode int
	Error      error
}

func (c *clMockCheckCVs) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clMockCheckCVs) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	var body []byte
	if strings.HasSuffix(url, "Name='ERROR'&Version='1.2.3'") {
		return &http.Response{
			StatusCode: c.StatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, errors.New("Validate went wrong")
	}
	body = []byte(responseCheckCVs)
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, c.Error
}

var responseCheckCVs = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/SoftwareComponentVersionSet(Name='%2FDRNMSPC%2FCOMP01',Version='0001')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/SoftwareComponentVersionSet(Name='%2FDRNMSPC%2FCOMP01',Version='0001')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.SoftwareComponentVersion"
        },
        "Name": "/DRNMSPC/COMP01",
        "Version": "0001",
        "SpLevel": "0002",
        "PatchLevel": "0003",
        "Vendor": "",
        "VendorType": ""
    }
}`
