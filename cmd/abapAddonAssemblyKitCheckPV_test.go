package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
)

func TestInitPV(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckPV{}
		prodvers := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "1.2.3",
		}

		var pv pv
		pv.init(prodvers, *conn)
		assert.Equal(t, "/DRNMSPC/PRD01", pv.Name)
		assert.Equal(t, "1.2.3", pv.VersionYAML)
	})
}

func TestValidatePV(t *testing.T) {
	t.Run("test validate", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckPV{}
		var pv pv
		pv.connector = *conn
		pv.Name = "/DRNMSPC/PRD01"
		pv.VersionYAML = "1.2.3"
		err := pv.validate()
		assert.NoError(t, err)
		assert.Equal(t, pv.Version, "0003")
		assert.Equal(t, pv.SpsLevel, "0002")
		assert.Equal(t, pv.PatchLevel, "0001")
	})
}

func TestValidatePVError(t *testing.T) {
	t.Run("test validate with error", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckPV{}
		var pv pv
		pv.connector = *conn
		pv.Name = "ERROR"
		pv.VersionYAML = "1.2.3"
		err := pv.validate()
		assert.Error(t, err)
		assert.Equal(t, pv.Version, "")
		assert.Equal(t, pv.SpsLevel, "")
		assert.Equal(t, pv.PatchLevel, "")
	})
}

func TestCopyFieldsPV(t *testing.T) {
	t.Run("test copyFieldsToRepo", func(t *testing.T) {
		prodVers := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "1.2.3",
		}
		var pv pv
		pv.Version = "0003"
		pv.SpsLevel = "0002"
		pv.PatchLevel = "0001"
		pv.copyFieldsToRepo(&prodVers)
		assert.Equal(t, "0003", prodVers.AddonVersion)
		assert.Equal(t, "0002", prodVers.AddonSpsLevel)
		assert.Equal(t, "0001", prodVers.AddonPatchLevel)
	})
}

type clMockCheckPV struct {
	StatusCode int
	Error      error
}

func (c *clMockCheckPV) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clMockCheckPV) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	var body []byte
	if strings.HasSuffix(url, "Name='ERROR'&Version='1.2.3'") {
		return &http.Response{
			StatusCode: c.StatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, errors.New("Validate went wrong")
	}
	body = []byte(responseCheckPV)
	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, c.Error
}

var responseCheckPV = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/ProductVersionSet(Name='%2FDRNMSPC%2FPRD01',Version='0001')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/ProductVersionSet(Name='%2FDRNMSPC%2FPRD01',Version='0001')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.ProductVersion"
        },
        "Name": "/DRNMSPC/PRD01",
        "Version": "0003",
        "SpsLevel": "0002",
        "PatchLevel": "0001",
        "Vendor": "",
        "VendorType": ""
    }
}`
