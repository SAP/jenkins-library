package build

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

// DownloadClientMock : Mock for Download Client used for artefact test
type DownloadClientMock struct{}

// DownloadFile : Empty file download
func (dc *DownloadClientMock) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return nil
}

// SetOptions : Download Client options
func (dc *DownloadClientMock) SetOptions(opts piperhttp.ClientOptions) {}

// ClMock : Mock for Build Framework Client used for BF test
type ClMock struct {
	Token      string
	StatusCode int
	Error      error
}

// SetOptions : BF Client options
func (c *ClMock) SetOptions(opts piperhttp.ClientOptions) {}

// SendRequest : BF Send Fake request
func (c *ClMock) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if method == "GET" || method == "POST" {
		body := []byte(fakeResponse(method, url))
		return &http.Response{
			StatusCode: c.StatusCode,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, c.Error
	} else if method == "HEAD" {
		var body []byte
		header := http.Header{}
		header.Set("X-Csrf-Token", c.Token)
		body = []byte("")
		return &http.Response{
			StatusCode: c.StatusCode,
			Header:     header,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, c.Error
	} else {
		return nil, c.Error
	}
}

func fakeResponse(method string, url string) string {
	if method == "POST" {
		return responsePOST
	}
	if strings.HasSuffix(url, "/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')") {
		return responseGetBuild
	} else if strings.HasSuffix(url, "/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')/tasks") {
		return responseGetTasks
	} else if strings.HasSuffix(url, "/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)/logs") {
		return responseGetLog0
	} else if strings.HasSuffix(url, "/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/logs") {
		return responseGetLog1
	} else if strings.HasSuffix(url, "/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')/values") {
		return responseGetValues
	} else if strings.HasSuffix(url, "tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)/results") {
		return responseGetResults0
	} else if strings.HasSuffix(url, "tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/results") {
		return responseGetResults1
	}
	return ""
}

var responseGetBuild = `{
	"d": {
		"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
		"run_state": "FINISHED",
		"result_state": "SUCCESSFUL",
		"phase": "test1",
		"entitytype": "P",
		"startedby": "SOME_USER",
		"started_at": "/Date(1591718108103+0000)/",
		"finished_at": "/Date(1591718129432+0000)/",
		"tasks": {
			"__deferred": {
				"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')/tasks"
			}
		},
		"values": {
			"__deferred": {
				"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')/values"
			}
		}
	}
}`

var responsePOST = `{
	"d": {
		"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
		"run_state": "ACCEPTED",
		"result_state": "",
		"phase": "test1",
		"entitytype": "",
		"startedby": "SOME_USER",
		"started_at": null,
		"finished_at": null,
		"tasks": {
			"__deferred": {
				"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPNVMOUQL2LHUFAUA')/tasks"
			}
		},
		"values": {
			"results": []
		}
	}
}`

var responseGetTasks = `{
	"d": {
		"results": [
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 1,
				"log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_1",
				"plugin_class": "/BUILD/CL_TEST_PLUGIN_OK",
				"started_at": "/Date(1591718128730+0000)/",
				"finished_at": "/Date(1591718129369+0000)/",
				"result_state": "SUCCESSFUL",
				"logs": {
					"__deferred": {
						"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/logs"
					}
				},
				"results": {
					"__deferred": {
						"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/results"
					}
				}
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 0,
				"log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_0",
				"plugin_class": "",
				"started_at": "/Date(1591718128728+0000)/",
				"finished_at": "/Date(1591718129462+0000)/",
				"result_state": "SUCCESSFUL",
				"logs": {
					"__deferred": {
						"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)/logs"
					}
				},
				"results": {
					"__deferred": {
						"uri": "https://some_server:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)/results"
					}
				}
			}
		]
	}
}`

var responseGetLog0 = `{
	"d": {
		"results": [
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 0,
				"log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_0",
				"msgty": "I",
				"detlevel": "3",
				"log_line": "I:/BUILD/LOG:000 ABAP Build Framework",
				"TIME_STMP": "20200721133523"
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 0,
				"log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_0",
				"msgty": "I",
				"detlevel": "3",
				"log_line": "I:/BUILD/LOG:000 ... Build Execution finished SUCCESSFUL",
				"TIME_STMP": "20200721133528"
			}
		]
	}
}`

var responseGetLog1 = `{
	"d": {
		"results": [
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 1,
				"log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_1",
				"msgty": "I",
				"detlevel": "1",
				"log_line": "I:/BUILD/LOG:000 Hello Packages [1]: , /BUILD/CORE, here is your lovely test_ok plugin!",
				"TIME_STMP": "20200721133528"
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 1,
				"log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_1",
				"msgty": "W",
				"detlevel": "3",
				"log_line": "W:/BUILD/LOG:000 We can even have warnings!",
				"TIME_STMP": "20200721133528"
			}
		]
	}
}`

var responseGetResults0 = `{
	"d": {
		"results": []
	}
}`

var responseGetResults1 = `{
	"d": {
		"results": [
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 1,
				"name": "HT-6111.JPG",
				"additional_info": "",
				"mimetype": "image/jpeg"
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"task_id": 1,
				"name": "SAR_XML",
				"additional_info": "/usr/sap/trans/tmp/SAPK-001AAINITAPC1.SAR",
				"mimetype": "application/octet-stream"
			}
		]
	}
}`

var responseGetValues = `{
	"d": {
		"results": [
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"value_id": "PHASE",
				"value": "test1"
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"value_id": "PACKAGES",
				"value": "/BUILD/CORE"
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"value_id": "season",
				"value": "winter"
			},
			{
				"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
				"value_id": "SUN",
				"value": "FLOWER"
			}
		]
	}
}`

func GetMockBuildTestDownloadPublish() Build {
	conn := new(Connector)
	conn.DownloadClient = &DownloadClientMock{}

	results0 := []Result{
		{
			connector: *conn,
			Name:      dummyResultName,
		},
	}
	results1 := []Result{
		{
			connector: *conn,
			Name:      "File1",
		},
		{
			connector: *conn,
			Name:      "File2",
		},
		{
			connector: *conn,
			Name:      "File3",
		},
	}

	build := Build{
		BuildID: "123",
		Tasks: []task{
			{
				TaskID:  0,
				Results: results0,
			},
			{
				TaskID:  1,
				Results: results1,
			},
		},
	}
	return build
}
