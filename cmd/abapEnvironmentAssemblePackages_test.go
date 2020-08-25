package cmd

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
)

func testSetup(client piperhttp.Sender, buildID string) build {
	conn := new(connector)
	conn.Client = client
	conn.DownloadClient = &downloadClientMock{}
	conn.Header = make(map[string][]string)
	b := build{
		connector: *conn,
		BuildID:   buildID,
	}
	return b
}

func TestCheckIfFailedAndPrintLogsWithError(t *testing.T) {
	t.Run("checkIfFailedAndPrintLogs with failed build", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		b.RunState = failed
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		err := checkIfFailedAndPrintLogs(buildsWithRepo)
		assert.Error(t, err)
	})
}

func TestCheckIfFailedAndPrintLogs(t *testing.T) {
	t.Run("checkIfFailedAndPrintLogs", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		b.RunState = finished
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		err := checkIfFailedAndPrintLogs(buildsWithRepo)
		assert.NoError(t, err)
	})
}

func TestStarting(t *testing.T) {
	t.Run("Run starting", func(t *testing.T) {
		client := &clMock{
			Token: "MyToken",
		}
		conn := new(connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:        "RepoA",
			Version:     "0001",
			PackageName: "Package",
			PackageType: "AOI",
			SpLevel:     "0000",
			PatchLevel:  "0000",
			Status:      "P",
			Namespace:   "/DEMO/",
		}
		repos = append(repos, repo)
		repo.Status = "R"
		repos = append(repos, repo)

		builds, buildsAlreadyReleased, err := starting(repos, *conn)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(builds))
		assert.Equal(t, 1, len(buildsAlreadyReleased))
		assert.Equal(t, accepted, builds[0].build.RunState)
		assert.Equal(t, runState(""), buildsAlreadyReleased[0].build.RunState)
	})
}

func TestStartingInvalidInput(t *testing.T) {
	t.Run("Run starting", func(t *testing.T) {
		client := &clMock{
			Token: "MyToken",
		}
		conn := new(connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:   "RepoA",
			Status: "P",
		}
		repos = append(repos, repo)
		_, _, err := starting(repos, *conn)
		assert.Error(t, err)
	})
}

func TestPolling(t *testing.T) {
	t.Run("Run polling", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		err := polling(buildsWithRepo, 600, 1)
		assert.NoError(t, err)
		assert.Equal(t, finished, buildsWithRepo[0].build.RunState)
	})
}

func TestDownloadSARXML(t *testing.T) {
	t.Run("Run downloadSARXML", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		repos, err := downloadSARXML(buildsWithRepo)
		assert.NoError(t, err)
		downloadPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment", "abap", "SAPK-001AAINITAPC1.SAR")
		assert.Equal(t, downloadPath, repos[0].SarXMLFilePath)
	})
}

// *******************************************************************************************************************************
// ************************************************* Tests for REUSE Part ********************************************************
// *******************************************************************************************************************************

func TestSTart(t *testing.T) {
	t.Run("Run start", func(t *testing.T) {
		client := &clMock{
			Token: "MyToken",
		}
		b := testSetup(client, "")
		inputValues := values{
			Values: []value{
				{
					ValueID: "PACKAGES",
					Value:   "/BUILD/CORE",
				},
				{
					ValueID: "season",
					Value:   "winter",
				},
			},
		}
		err := b.start("test", inputValues)
		assert.NoError(t, err)
		assert.Equal(t, accepted, b.RunState)
	})
}

func TestGet(t *testing.T) {
	t.Run("Run Get", func(t *testing.T) {
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		err := b.get()
		assert.NoError(t, err)
		assert.Equal(t, finished, b.RunState)
		assert.Equal(t, 0, len(b.Tasks))
	})
}

func TestGetTasks(t *testing.T) {
	t.Run("Run getTasks", func(t *testing.T) {
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		assert.Equal(t, 0, len(b.Tasks))
		err := b.getTasks()
		assert.NoError(t, err)
		assert.Equal(t, b.Tasks[0].TaskID, 0)
		assert.Equal(t, b.Tasks[0].PluginClass, "")
		assert.Equal(t, b.Tasks[1].TaskID, 1)
		assert.Equal(t, b.Tasks[1].PluginClass, "/BUILD/CL_TEST_PLUGIN_OK")
	})
}

func TestGetLogs(t *testing.T) {
	t.Run("Run getLogs", func(t *testing.T) {
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		err := b.getLogs()
		assert.NoError(t, err)
		assert.Equal(t, "I:/BUILD/LOG:000 ABAP Build Framework", b.Tasks[0].Logs[0].Logline)
		assert.Equal(t, loginfo, b.Tasks[0].Logs[0].Msgty)
		assert.Equal(t, "W:/BUILD/LOG:000 We can even have warnings!", b.Tasks[1].Logs[1].Logline)
		assert.Equal(t, logwarning, b.Tasks[1].Logs[1].Msgty)
	})
}

func TestGetValues(t *testing.T) {
	t.Run("Run getValues", func(t *testing.T) {
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		assert.Equal(t, 0, len(b.Values))
		err := b.getValues()
		assert.NoError(t, err)
		assert.Equal(t, 4, len(b.Values))
		assert.Equal(t, "PHASE", b.Values[0].ValueID)
		assert.Equal(t, "test1", b.Values[0].Value)
		assert.Equal(t, "PACKAGES", b.Values[1].ValueID)
		assert.Equal(t, "/BUILD/CORE", b.Values[1].Value)
		assert.Equal(t, "season", b.Values[2].ValueID)
		assert.Equal(t, "winter", b.Values[2].Value)
		assert.Equal(t, "SUN", b.Values[3].ValueID)
		assert.Equal(t, "FLOWER", b.Values[3].Value)
	})
}

func TestGetResults(t *testing.T) {
	t.Run("Run getResults", func(t *testing.T) {
		b := testSetup(&clMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		err := b.getResults()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(b.Tasks[0].Results))
		assert.Equal(t, 2, len(b.Tasks[1].Results))
		assert.Equal(t, "image/jpeg", b.Tasks[1].Results[0].Mimetype)
		assert.Equal(t, "application/octet-stream", b.Tasks[1].Results[1].Mimetype)

		_, err = b.getResult("does_not_exist")
		assert.Error(t, err)
		r, err := b.getResult("SAR_XML")
		assert.Equal(t, "application/octet-stream", r.Mimetype)
		assert.NoError(t, err)
	})
}

type downloadClientMock struct{}

func (dc *downloadClientMock) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return nil
}

func (dc *downloadClientMock) SetOptions(opts piperhttp.ClientOptions) {}

type clMock struct {
	Token      string
	StatusCode int
	Error      error
}

func (c *clMock) SetOptions(opts piperhttp.ClientOptions) {}

func (c *clMock) SendRequest(method string, url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if method == "GET" || method == "POST" {
		var body []byte
		body = []byte(fakeResponse(method, url))
		return &http.Response{
			StatusCode: c.StatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, c.Error
	} else if method == "HEAD" {
		var body []byte
		header := http.Header{}
		header.Set("X-Csrf-Token", c.Token)
		body = []byte("")
		return &http.Response{
			StatusCode: c.StatusCode,
			Header:     header,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
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
		return ResponseGetLog0
	} else if strings.HasSuffix(url, "/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/logs") {
		return ResponseGetLog1
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
		"__metadata": {
			"id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')",
			"uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')",
			"type": "BUILD.CORE_SRV.xBUILDxVIEW_BUILDSType"
		},
		"build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
		"run_state": "FINISHED",
		"result_state": "SUCCESSFUL",
		"phase": "test1",
		"entitytype": "P",
		"startedby": "BENTELER",
		"started_at": "/Date(1591718108103+0000)/",
		"finished_at": "/Date(1591718129432+0000)/",
		"tasks": {
			"__deferred": {
				"uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')/tasks"
			}
		},
		"values": {
			"__deferred": {
				"uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPOVMXK4DNPBDRW2M')/values"
			}
		}
	}
}`

var responsePOST = `{
    "d": {
        "__metadata": {
            "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPNVMOUQL2LHUFAUA')",
            "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPNVMOUQL2LHUFAUA')",
            "type": "BUILD.CORE_SRV.xBUILDxVIEW_BUILDSType"
        },
        "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
        "run_state": "ACCEPTED",
        "result_state": "",
        "phase": "test1",
        "entitytype": "",
        "startedby": "BENTELER",
        "started_at": null,
        "finished_at": null,
        "tasks": {
            "__deferred": {
                "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/builds('ABIFNLDCSQPNVMOUQL2LHUFAUA')/tasks"
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
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_TASKSType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "task_id": 1,
                "log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_1",
                "plugin_class": "/BUILD/CL_TEST_PLUGIN_OK",
                "started_at": "/Date(1591718128730+0000)/",
                "finished_at": "/Date(1591718129369+0000)/",
                "result_state": "SUCCESSFUL",
                "logs": {
                    "__deferred": {
                        "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/logs"
                    }
                },
                "results": {
                    "__deferred": {
                        "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1)/results"
                    }
                }
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_TASKSType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "task_id": 0,
                "log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_0",
                "plugin_class": "",
                "started_at": "/Date(1591718128728+0000)/",
                "finished_at": "/Date(1591718129462+0000)/",
                "result_state": "SUCCESSFUL",
                "logs": {
                    "__deferred": {
                        "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)/logs"
                    }
                },
                "results": {
                    "__deferred": {
                        "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0)/results"
                    }
                }
            }
        ]
    }
}`

var ResponseGetLog0 = `{
    "d": {
        "results": [
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_0')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_0')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_LOGSType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "task_id": 0,
                "log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_0",
                "msgty": "I",
                "detlevel": "3",
                "log_line": "I:/BUILD/LOG:000 ABAP Build Framework",
                "TIME_STMP": "20200721133523"
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_0')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=0,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_0')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_LOGSType"
                },
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

var ResponseGetLog1 = `{
    "d": {
        "results": [
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_1')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_1')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_LOGSType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "task_id": 1,
                "log_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M_1",
                "msgty": "I",
                "detlevel": "1",
                "log_line": "I:/BUILD/LOG:000 Hello Packages [1]: , /BUILD/CORE, here is your lovely test_ok plugin!",
                "TIME_STMP": "20200721133528"
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_1')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/logs(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,log_id='ABIFNLDCSQPOVMXK4DNPBDRW2M_1')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_LOGSType"
                },
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
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/results(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,name='HT-6111.JPG')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/results(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,name='HT-6111.JPG')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_RESULTSType",
                    "content_type": "image/jpeg",
                    "media_src": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/results(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,name='HT-6111.JPG')/$value"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "task_id": 1,
                "name": "HT-6111.JPG",
                "additional_info": "",
                "mimetype": "image/jpeg"
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/results(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,name='2times_hello')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/results(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,name='2times_hello')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_RESULTSType",
                    "content_type": "text/plain",
                    "media_src": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/results(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',task_id=1,name='2times_hello')/$value"
                },
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
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='PHASE')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='PHASE')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_VALUESType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "value_id": "PHASE",
                "value": "test1"
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='PACKAGES')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='PACKAGES')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_VALUESType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "value_id": "PACKAGES",
                "value": "/BUILD/CORE"
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='season')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='season')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_VALUESType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "value_id": "season",
                "value": "winter"
            },
            {
                "__metadata": {
                    "id": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='SUN')",
                    "uri": "https://ldai3yi3.wdf.sap.corp:44334/sap/opu/odata/BUILD/CORE_SRV/values(build_id='ABIFNLDCSQPOVMXK4DNPBDRW2M',value_id='SUN')",
                    "type": "BUILD.CORE_SRV.xBUILDxVIEW_VALUESType"
                },
                "build_id": "ABIFNLDCSQPOVMXK4DNPBDRW2M",
                "value_id": "SUN",
                "value": "FLOWER"
            }
        ]
    }
}`
