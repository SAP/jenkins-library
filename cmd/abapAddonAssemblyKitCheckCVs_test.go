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

func TestCheckCVs(t *testing.T) {
	t.Run("test init, validate and addFields", func(t *testing.T) {
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
		err := c.validate()
		assert.NoError(t, err)
		c.copyFieldsToRepo(&repo)
		assert.Equal(t, "0001", repo.Version)
		assert.Equal(t, "0002", repo.SpLevel)
		assert.Equal(t, "0003", repo.PatchLevel)
	})
}

type clMockCheckCVs struct {
	StatusCode int
	Error      error
}

func (c *clMockCheckCVs) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clMockCheckCVs) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	var body []byte
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
