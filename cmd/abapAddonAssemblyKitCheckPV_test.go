package cmd

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
)

func TestCheckPV(t *testing.T) {
	t.Run("test init, validate and addFields", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockCheckPV{}
		addonDescriptor := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "3.2.1",
		}
		var p pv
		p.init(addonDescriptor, *conn)
		assert.Equal(t, "/DRNMSPC/PRD01", p.Name)
		assert.Equal(t, "3.2.1", p.VersionYAML)
		err := p.validate()
		assert.NoError(t, err)
		p.copyFieldsToRepo(&addonDescriptor)
		assert.Equal(t, "0003", addonDescriptor.AddonVersion)
		assert.Equal(t, "0002", addonDescriptor.AddonSpsLevel)
		assert.Equal(t, "0001", addonDescriptor.AddonPatchLevel)
	})
}

type clMockCheckPV struct {
	StatusCode int
	Error      error
}

func (c *clMockCheckPV) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clMockCheckPV) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	var body []byte
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
