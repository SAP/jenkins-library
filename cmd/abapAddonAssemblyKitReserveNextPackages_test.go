package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// ********************* Test init *******************
func TestInitPackage(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &clMockReservePackages{}
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.0.0",
		}
		var p pckg
		p.init(repo, *conn)
		assert.Equal(t, "/DRNMSPC/COMP01", p.ComponentName)
		assert.Equal(t, "1.0.0", p.VersionYAML)
	})
}

// ********************* Test copyFieldsToRepositories *******************
func TestCopyFieldsToRepositoriesPackage(t *testing.T) {
	t.Run("test copyFieldsToRepositories", func(t *testing.T) {
		pckgWR := []packageWithRepository{
			{
				p: pckg{
					ComponentName: "/DRNMSPC/COMP01",
					VersionYAML:   "1.0.0",
					PackageName:   "SAPK-001AAINDRNMSPC",
					Type:          "AOI",
					Status:        planned,
					Namespace:     "/DRNMSPC/",
				},
				repo: abaputils.Repository{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.0.0",
				},
			},
		}
		repos := copyFieldsToRepositories(pckgWR)
		assert.Equal(t, "SAPK-001AAINDRNMSPC", repos[0].PackageName)
		assert.Equal(t, "AOI", repos[0].PackageType)
		assert.Equal(t, string(planned), repos[0].Status)
		assert.Equal(t, "/DRNMSPC/", repos[0].Namespace)
	})
}

// ********************* Test reserveNext *******************
func TestReserveNextPackage(t *testing.T) {
	t.Run("test reserveNext", func(t *testing.T) {
		p := testPackageSetup("/DRNMSPC/COMP01", "1.0.0")
		err := p.reserveNext()
		assert.NoError(t, err)
		assert.Equal(t, "SAPK-001AAINDRNMSPC", p.PackageName)
		assert.Equal(t, "AOI", p.Type)
		assert.Equal(t, planned, p.Status)
	})
}

func TestReserveNextInvalidInputPackage(t *testing.T) {
	t.Run("test reserveNext missing versionYAML", func(t *testing.T) {
		p := testPackageSetup("/DRNMSPC/COMP01", "")
		err := p.reserveNext()
		assert.Error(t, err)
		assert.Equal(t, "", p.PackageName)
		assert.Equal(t, "", p.Type)
		assert.Equal(t, packageStatus(""), p.Status)
	})
}

func TestReserveNextResponseErrorPackage(t *testing.T) {
	t.Run("test reserveNext", func(t *testing.T) {
		p := testPackageSetup("ERROR", "1.0.0")
		err := p.reserveNext()
		assert.Error(t, err)
		assert.Equal(t, "", p.PackageName)
		assert.Equal(t, "", p.Type)
		assert.Equal(t, packageStatus(""), p.Status)
	})
}

// ********************* Test reservePackages *******************

func TestReservePackages(t *testing.T) {
	t.Run("test reservePackages", func(t *testing.T) {
		repositories, conn := testRepositoriesSetup("/DRNMSPC/COMP01", "1.0.0")
		repos, err := reservePackages(repositories, conn)
		assert.NoError(t, err)
		assert.Equal(t, "/DRNMSPC/COMP01", repos[0].p.ComponentName)
		assert.Equal(t, "1.0.0", repos[0].p.VersionYAML)
		assert.Equal(t, planned, repos[0].p.Status)
	})
}

func TestReservePackagesError(t *testing.T) {
	t.Run("test reservePackages with error", func(t *testing.T) {
		repositories, conn := testRepositoriesSetup("ERROR", "1.0.0")
		_, err := reservePackages(repositories, conn)
		assert.Error(t, err)
	})
}

// ********************* Test pollReserveNextPackages *******************

func TestPollReserveNextPackages(t *testing.T) {
	t.Run("test pollReserveNextPackages testing loop", func(t *testing.T) {
		pckgWR := testPollPackagesSetup(planned)
		err := pollReserveNextPackages(pckgWR, 5, 1)
		assert.NoError(t, err)
		assert.Equal(t, planned, pckgWR[0].p.Status)
		assert.Equal(t, "/DRNMSPC/", pckgWR[0].p.Namespace)
	})
}

func TestPollReserveNextLocked(t *testing.T) {
	t.Run("test pollReserveNextPackages status locked", func(t *testing.T) {
		pckgWR := testPollPackagesSetup(locked)
		err := pollReserveNextPackages(pckgWR, 5, 1)
		assert.Error(t, err)
		assert.Equal(t, locked, pckgWR[0].p.Status)
	})
}

func TestPollReserveNextReleased(t *testing.T) {
	t.Run("test pollReserveNextPackages status released", func(t *testing.T) {
		pckgWR := testPollPackagesSetup(released)
		err := pollReserveNextPackages(pckgWR, 5, 1)
		assert.NoError(t, err)
		assert.Equal(t, released, pckgWR[0].p.Status)
	})
}

func TestPollReserveNextUnknownStatus(t *testing.T) {
	t.Run("test pollReserveNextPackages unknow status", func(t *testing.T) {
		pckgWR := testPollPackagesSetup("X")
		err := pollReserveNextPackages(pckgWR, 5, 1)
		assert.Error(t, err)
		assert.Equal(t, packageStatus("X"), pckgWR[0].p.Status)
	})
}

func TestPollReserveNextTimeout(t *testing.T) {
	t.Run("test pollReserveNextPackages testing timeout", func(t *testing.T) {
		pckgWR := testPollPackagesSetup("timeout")
		var timeout time.Duration
		timeout = 2 * time.Second
		err := pollReserveNextPackages(pckgWR, timeout, 1)
		assert.Error(t, err)
	})
}

// ********************* Setup functions *******************

func testPollPackagesSetup(finalState packageStatus) []packageWithRepository {
	conn := new(connector)
	conn.Client = &clMockReservePackages{
		finalState: finalState,
	}
	conn.Header = make(map[string][]string)
	pckgWR := []packageWithRepository{
		{
			p: pckg{
				connector:     *conn,
				ComponentName: "/DRNMSPC/COMP01",
				VersionYAML:   "1.0.0",
				PackageName:   "SAPK-001AAINDRNMSPC",
				Type:          "AOI",
			},
			repo: abaputils.Repository{},
		},
	}
	return pckgWR
}

func testRepositoriesSetup(componentName string, versionYAML string) ([]abaputils.Repository, connector) {
	conn := new(connector)
	conn.Client = &clMockReservePackages{}
	conn.Header = make(map[string][]string)
	repositories := []abaputils.Repository{
		{
			Name:        componentName,
			VersionYAML: versionYAML,
		},
	}
	return repositories, *conn
}

func testPackageSetup(componentName string, versionYAML string) pckg {
	conn := new(connector)
	conn.Client = &clMockReservePackages{}
	conn.Header = make(map[string][]string)
	p := pckg{
		connector:     *conn,
		ComponentName: componentName,
		VersionYAML:   versionYAML,
	}
	return p
}

// ********************* Mocking *******************

type clMockReservePackages struct {
	finalState packageStatus
	counter    int
}

func (c *clMockReservePackages) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clMockReservePackages) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	switch method {
	case "HEAD":
		return c.sendRequestHead()
	case "POST":
		return c.sendRequestPost(url)
	case "GET":
		return c.sendRequestGet()
	}
	return nil, nil
}

func (c *clMockReservePackages) sendRequestHead() (*http.Response, error) {
	var body []byte
	header := http.Header{}
	header.Set("X-CSRF-Token", "myToken")
	body = []byte("")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (c *clMockReservePackages) sendRequestPost(url string) (*http.Response, error) {
	var body []byte
	if strings.HasSuffix(url, "Name='ERROR'&Version='1.0.0'") {
		return &http.Response{
			StatusCode: 400,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, errors.New("reserveNext went wrong")
	}
	body = []byte(responseReserveNextPackagePost)
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (c *clMockReservePackages) sendRequestGet() (*http.Response, error) {
	var body []byte
	var err error
	switch c.finalState {
	case planned:
		c.counter++
		switch c.counter {
		case 1:
			err = errors.New("get went wrong")
			// body = []byte("")
		case 2:
			body = []byte(responseReserveNextPackageCreationTriggered)
		case 3:
			body = []byte(responseReserveNextPackagePlanned)
		}
	case locked:
		body = []byte(responseReserveNextPackageLocked)
	case released:
		body = []byte(responseReserveNextPackageReleased)
	case "X":
		body = []byte(responseReserveNextPackageUnknownState)
	case "timeout":
		body = []byte(responseReserveNextPackageCreationTriggered)
	}
	if err != nil {
		return &http.Response{
			StatusCode: 400,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, err
	}
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, nil
}

// ********************* Testdata *******************

var responseReserveNextPackagePost = `{
    "d": {
        "DeterminePackageForScv": {
            "__metadata": {
                "type": "SSDA.AAS_ODATA_PACKAGE_SRV.PackageExtended"
            },
            "Name": "SAPK-001AAINDRNMSPC",
            "Type": "AOI",
            "ScName": "/DRNMSPC/COMP01",
            "ScVersion": "0001",
            "SpLevel": "0000",
            "PatchLevel": "0000",
            "Predecessor": "",
            "PredecessorCommitId": "",
            "Status": "P"
        }
    }
}`

var responseReserveNextPackageCreationTriggered = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "C",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackageLocked = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "L",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackagePlanned = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "P",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackageReleased = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "R",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackageUnknownState = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "X",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`
