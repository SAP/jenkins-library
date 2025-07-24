package build

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

// MockClient : use NewMockClient for construction
type MockClient struct {
	//Key = HTTP-Method + Url
	Data map[string][]http.Response
}

// MockData: data for the mockClient
type MockData struct {
	Method     string
	Url        string
	Body       string
	StatusCode int
	Header     http.Header
}

// NewMockClient : Constructs a new Mock Client implementing piperhttp.Sender
func NewMockClient() MockClient {
	var ret = MockClient{}
	ret.Data = make(map[string][]http.Response)
	return ret
}

// AddResponse : adds a response object to the mock lib
func (mc *MockClient) AddResponse(Method, Url string, response http.Response) {
	responseList, ok := mc.Data[Method+Url]
	if !ok {
		responseList = make([]http.Response, 0)
	}
	responseList = append(responseList, response)

	mc.Data[Method+Url] = responseList
}

// Add : adds a response with the given Body and statusOK to the mock lib
func (mc *MockClient) Add(Method, Url, Body string) {
	mc.AddResponse(Method, Url, http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewReader([]byte(Body))),
	})
}

// AddBody : adds a response with the given data to the mock lib
func (mc *MockClient) AddBody(Method, Url, Body string, StatusCode int, header http.Header) {
	mc.AddResponse(Method, Url, http.Response{
		StatusCode: StatusCode,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader([]byte(Body))),
	})
}

// AddData : add the mock Data as response to the mock lib
func (mc *MockClient) AddData(data MockData) {
	mc.AddResponse(data.Method, data.Url, http.Response{
		StatusCode: data.StatusCode,
		Header:     data.Header,
		Body:       io.NopCloser(bytes.NewReader([]byte(data.Body))),
	})
}

func (mc *MockClient) getResponse(Method, Url string) (http.Response, bool) {
	responseList, ok := mc.Data[Method+Url]
	if !ok {
		return http.Response{}, false
	}
	response := responseList[0]
	if len(responseList) > 1 {
		mc.Data[Method+Url] = responseList[1:]
	} else {
		delete(mc.Data, Method+Url)
	}

	return response, true
}

// SetOptions : dummy as of now
func (mc *MockClient) SetOptions(opts piperhttp.ClientOptions) {}

// SendRequest sets a HTTP response for a client mock
func (mc *MockClient) SendRequest(Method, Url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	response, ok := mc.getResponse(Method, Url)
	if !ok {
		return nil, fmt.Errorf("No Mock data for %s", Method+Url)
	}
	return &response, nil
}

// DownloadFile : Empty file download
func (mc *MockClient) DownloadFile(Url, filename string, header http.Header, cookies []*http.Cookie) error {
	return nil
}

/***************************************
*** BuildMock
***************************************/

// GetBuildMockClient : Constructs a Mock Client with example build Requests/Responses
func GetBuildMockClient() MockClient {
	mc := NewMockClient()

	mc.AddData(buildHead)
	mc.AddData(buildPost)
	mc.AddData(buildGet1)
	mc.AddData(buildGet2)
	mc.AddData(buildGetTasks)
	mc.AddData(buildGetTask0Logs)
	mc.AddData(buildGetTask1Logs)
	mc.AddData(buildGetTask2Logs)
	mc.AddData(buildGetTask3Logs)
	mc.AddData(buildGetTask4Logs)
	mc.AddData(buildGetTask5Logs)
	mc.AddData(buildGetTask6Logs)
	mc.AddData(buildGetTask7Logs)
	mc.AddData(buildGetTask8Logs)
	mc.AddData(buildGetTask9Logs)
	mc.AddData(buildGetTask10Logs)
	mc.AddData(buildGetTask11Logs)
	mc.AddData(buildGetTask12Logs)
	mc.AddData(buildGetTask0Result)
	mc.AddData(buildGetTask1Result)
	mc.AddData(buildGetTask2Result)
	mc.AddData(buildGetTask3Result)
	mc.AddData(buildGetTask4Result)
	mc.AddData(buildGetTask5Result)
	mc.AddData(buildGetTask6Result)
	mc.AddData(buildGetTask7Result)
	mc.AddData(buildGetTask8Result)
	mc.AddData(buildGetTask9Result)
	mc.AddData(buildGetTask10Result)
	mc.AddData(buildGetTask11Result)
	mc.AddData(buildGetTask12Result)
	mc.AddData(buildGetTask11ResultMedia)
	mc.AddData(buildGetValues)

	return mc
}

func GetBuildMockClientWithClient() MockClient {
	mc := NewMockClient()
	mc.AddData(buildHeadWithClient)
	mc.AddData(buildPostWithClient)

	mc.AddData(buildGetTasksWithClient)

	mc.AddData(buildGet2WithClient)
	mc.AddData(buildGetTask0LogsWithClient)
	mc.AddData(buildGetValuesWithClient)
	return mc
}

// GetBuildMockClientToRun2Times : Constructs a Mock Client with example build Requests/Responses, this can run two times
func GetBuildMockClientToRun2Times() MockClient {
	mc := NewMockClient()

	mc.AddData(buildHead)
	mc.AddData(buildHead)

	mc.AddData(buildPost)
	mc.AddData(buildPost)

	mc.AddData(buildGet1)
	mc.AddData(buildGet2)
	mc.AddData(buildGet1)
	mc.AddData(buildGet2)

	mc.AddData(buildGetTasks)
	mc.AddData(buildGetTasks)

	mc.AddData(buildGetTask0Logs)
	mc.AddData(buildGetTask1Logs)
	mc.AddData(buildGetTask2Logs)
	mc.AddData(buildGetTask3Logs)
	mc.AddData(buildGetTask4Logs)
	mc.AddData(buildGetTask5Logs)
	mc.AddData(buildGetTask6Logs)
	mc.AddData(buildGetTask7Logs)
	mc.AddData(buildGetTask8Logs)
	mc.AddData(buildGetTask9Logs)
	mc.AddData(buildGetTask10Logs)
	mc.AddData(buildGetTask11Logs)
	mc.AddData(buildGetTask12Logs)

	mc.AddData(buildGetTask0Logs)
	mc.AddData(buildGetTask1Logs)
	mc.AddData(buildGetTask2Logs)
	mc.AddData(buildGetTask3Logs)
	mc.AddData(buildGetTask4Logs)
	mc.AddData(buildGetTask5Logs)
	mc.AddData(buildGetTask6Logs)
	mc.AddData(buildGetTask7Logs)
	mc.AddData(buildGetTask8Logs)
	mc.AddData(buildGetTask9Logs)
	mc.AddData(buildGetTask10Logs)
	mc.AddData(buildGetTask11Logs)
	mc.AddData(buildGetTask12Logs)

	mc.AddData(buildGetTask0Result)
	mc.AddData(buildGetTask1Result)
	mc.AddData(buildGetTask2Result)
	mc.AddData(buildGetTask3Result)
	mc.AddData(buildGetTask4Result)
	mc.AddData(buildGetTask5Result)
	mc.AddData(buildGetTask6Result)
	mc.AddData(buildGetTask7Result)
	mc.AddData(buildGetTask8Result)
	mc.AddData(buildGetTask9Result)
	mc.AddData(buildGetTask10Result)
	mc.AddData(buildGetTask11Result)
	mc.AddData(buildGetTask12Result)

	mc.AddData(buildGetTask0Result)
	mc.AddData(buildGetTask1Result)
	mc.AddData(buildGetTask2Result)
	mc.AddData(buildGetTask3Result)
	mc.AddData(buildGetTask4Result)
	mc.AddData(buildGetTask5Result)
	mc.AddData(buildGetTask6Result)
	mc.AddData(buildGetTask7Result)
	mc.AddData(buildGetTask8Result)
	mc.AddData(buildGetTask9Result)
	mc.AddData(buildGetTask10Result)
	mc.AddData(buildGetTask11Result)
	mc.AddData(buildGetTask12Result)

	mc.AddData(buildGetTask11ResultMedia)
	mc.AddData(buildGetTask11ResultMedia)

	mc.AddData(buildGetValues)
	mc.AddData(buildGetValues)

	return mc
}

var buildHead = MockData{
	Method: `HEAD`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV`,
	Body: `<?xml version="1.0"?>
	<HTTP_Body/>`,
	StatusCode: 200,
	Header:     http.Header{"x-csrf-token": {"HRfJP0OhB9C9mHs2RRqUzw=="}},
}

var buildHeadWithClient = MockData{
	Method:     `HEAD`,
	Url:        `/sap/opu/odata/BUILD/CORE_SRV?sap-client=001`,
	Body:       buildHead.Body,
	StatusCode: 200,
	Header:     buildHead.Header,
}

var buildPost = MockData{
	Method: `POST`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds`,
	Body: `{
	"d" : {
		"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
		"run_state" : "ACCEPTED",
		"result_state" : "",
		"phase" : "BUILD_AOI",
		"entitytype" : "",
		"startedby" : "CC0000000001",
		"started_at" : null,
		"finished_at" : null,
		"tasks" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/tasks"
			}
		},
		"values" : {
			"results" : [
			]
		}
	}
}`,
	StatusCode: 201,
}

var buildPostWithClient = MockData{
	Method:     `POST`,
	Url:        `/sap/opu/odata/BUILD/CORE_SRV/builds?sap-client=001`,
	Body:       buildPost.Body,
	StatusCode: 201,
}

var buildGet1 = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')`,
	Body: `{
	"d" : {
		"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
		"run_state" : "RUNNING",
		"result_state" : "SUCCESSFUL",
		"phase" : "BUILD_AOI",
		"entitytype" : "C",
		"startedby" : "CC0000000001",
		"started_at" : "\/Date(1614108520862+0000)\/",
		"finished_at" : "\/Date(1614108535350+0000)\/",
		"tasks" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/tasks"
			}
		},
		"values" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/values"
			}
		}
	}
}`,
	StatusCode: 200,
}

var buildGet2 = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')`,
	Body: `{
	"d" : {
		"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
		"run_state" : "FINISHED",
		"result_state" : "SUCCESSFUL",
		"phase" : "BUILD_AOI",
		"entitytype" : "C",
		"startedby" : "CC0000000001",
		"started_at" : "\/Date(1614108520862+0000)\/",
		"finished_at" : "\/Date(1614108535350+0000)\/",
		"tasks" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/tasks"
			}
		},
		"values" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/values"
			}
		}
	}
}`,
	StatusCode: 200,
}

var buildGet2WithClient = MockData{
	Method:     `GET`,
	Url:        `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')?sap-client=001`,
	Body:       buildGet2.Body,
	StatusCode: 200,
}

var buildGetRunStateFailed = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')`,
	Body: `{
	"d" : {
		"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
		"run_state" : "FAILED",
		"result_state" : "SUCCESSFUL",
		"phase" : "BUILD_AOI",
		"entitytype" : "C",
		"startedby" : "CC0000000001",
		"started_at" : "\/Date(1614108520862+0000)\/",
		"finished_at" : "\/Date(1614108535350+0000)\/",
		"tasks" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/tasks"
			}
		},
		"values" : {
			"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/values"
			}
		}
	}
}`,
	StatusCode: 200,
}

var buildGetTasks = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/tasks`,
	Body: `{
	"d" : {
		"results" : [
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 1,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_1",
			"plugin_class" : "CL_CD_SDC_SET_CVERS",
			"started_at" : "\/Date(1614108521633+0000)\/",
			"finished_at" : "\/Date(1614108521637+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=1)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=1)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 2,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_2",
			"plugin_class" : "/BUILD/GIT_COMMITS_2_PIECELIST",
			"started_at" : "\/Date(1614108521651+0000)\/",
			"finished_at" : "\/Date(1614108521864+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=2)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=2)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 3,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_3",
			"plugin_class" : "CL_CD_SDC_SET_TRDELVCHK",
			"started_at" : "\/Date(1614108521869+0000)\/",
			"finished_at" : "\/Date(1614108521889+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=3)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=3)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 4,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_4",
			"plugin_class" : "CL_CD_SDC_CREATE_UPINS",
			"started_at" : "\/Date(1614108521896+0000)\/",
			"finished_at" : "\/Date(1614108522293+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=4)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=4)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 5,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_5",
			"plugin_class" : "CL_CD_SDC_MERGE_BY_LIST",
			"started_at" : "\/Date(1614108522300+0000)\/",
			"finished_at" : "\/Date(1614108523039+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=5)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=5)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 6,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_6",
			"plugin_class" : "CL_CD_SDC_RELEASE_AOU",
			"started_at" : "\/Date(1614108523046+0000)\/",
			"finished_at" : "\/Date(1614108523230+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=6)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=6)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 7,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_7",
			"plugin_class" : "CL_CD_SDC_MERGE_AOI",
			"started_at" : "\/Date(1614108523236+0000)\/",
			"finished_at" : "\/Date(1614108523661+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=7)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=7)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 8,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_8",
			"plugin_class" : "CL_CD_SDC_ADD_REQUIRED_OBJECTS",
			"started_at" : "\/Date(1614108523668+0000)\/",
			"finished_at" : "\/Date(1614108523686+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=8)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=8)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 9,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_9",
			"plugin_class" : "CL_CD_SDC_CHECK_FT",
			"started_at" : "\/Date(1614108523692+0000)\/",
			"finished_at" : "\/Date(1614108524528+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=9)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=9)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 10,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_10",
			"plugin_class" : "CL_CD_SDC_RELEASE_FT",
			"started_at" : "\/Date(1614108524534+0000)\/",
			"finished_at" : "\/Date(1614108534847+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=10)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=10)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 11,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_11",
			"plugin_class" : "CL_CD_SDC_ASSEMBLE_FT",
			"started_at" : "\/Date(1614108534855+0000)\/",
			"finished_at" : "\/Date(1614108535145+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=11)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=11)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 12,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_12",
			"plugin_class" : "CL_CD_SDC_GET_DELIVERY_LOGS_FT",
			"started_at" : "\/Date(1614108535152+0000)\/",
			"finished_at" : "\/Date(1614108535343+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=12)/logs"
				}
			},
			"results" : {
				"__deferred" : {
				"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=12)/results"
				}
			}
		},
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 0,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
			"plugin_class" : "",
			"started_at" : "\/Date(1614108521633+0000)\/",
			"finished_at" : "\/Date(1614108535356+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/results"
				}
			}
		}
	]
	}
}`,
	StatusCode: 200,
}

var buildGetTasksWithClient = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/tasks?sap-client=001`,
	Body: `{
	"d" : {
		"results" : [
		{
			"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
			"task_id" : 0,
			"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
			"plugin_class" : "",
			"started_at" : "\/Date(1614108521633+0000)\/",
			"finished_at" : "\/Date(1614108535356+0000)\/",
			"result_state" : "SUCCESSFUL",
			"logs" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/logs"
				}
			},
			"results" : {
				"__deferred" : {
					"uri" : "https://some_server/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/results"
				}
			}
		}
	]
	}
}`,
	StatusCode: 200,
}

var buildGetTask0Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/logs`,
	Body: `{
		"d" : {
			"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 ⌂⌂⌂ ABAP Build Framework ⌂⌂⌂",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 ============================",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build 'AKO22FYOFYPOXHOBVKXUTX3A3Q' for execution of Phase 'BUILD_AOI' started",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Start Tree Initialization...",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Initializing build tree for software component /ITAPC1/I_CURRENCY",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Package Tree contains 2 packages",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 # of computed Build Packages: 1",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 # of computed Build Tasks: 12",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 ...Tree initialization finished",
				"TIME_STMP" : "20210223192841"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Start Build Execution...",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 1 plugin CL_CD_SDC_SET_CVERS executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 2 plugin /BUILD/GIT_COMMITS_2_PIECELIST executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 3 plugin CL_CD_SDC_SET_TRDELVCHK executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 4 plugin CL_CD_SDC_CREATE_UPINS executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 5 plugin CL_CD_SDC_MERGE_BY_LIST executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192843"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 6 plugin CL_CD_SDC_RELEASE_AOU executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192843"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 7 plugin CL_CD_SDC_MERGE_AOI executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192844"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 8 plugin CL_CD_SDC_ADD_REQUIRED_OBJECTS executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192844"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 9 plugin CL_CD_SDC_CHECK_FT executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192845"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 10 plugin CL_CD_SDC_RELEASE_FT executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192855"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 11 plugin CL_CD_SDC_ASSEMBLE_FT executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192855"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Build Task 12 plugin CL_CD_SDC_GET_DELIVERY_LOGS_FT executed SUCCESSFUL for scope Packages [2]: , /ITAPC1/I_CURRENCY, /ITAPC1/I_CURRENCY_DEV",
				"TIME_STMP" : "20210223192855"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 0,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_0",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 ... Build Execution finished SUCCESSFUL",
				"TIME_STMP" : "20210223192855"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask0LogsWithClient = MockData{
	Method:     `GET`,
	Url:        `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/logs?sap-client=001`,
	Body:       buildGetTask0Logs.Body,
	StatusCode: 200,
}

var buildGetTask1Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=1)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 1,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_1",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 CVERS values set: /ITAPC1/I_CURRENCY, 0001, 0000",
				"TIME_STMP" : "20210223192842"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask2Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=2)/logs`,
	Body: `{
	"d" : {
		"results" : [
			]
	}
}`,
	StatusCode: 200,
}

var buildGetTask3Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=3)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 3,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_3",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 TRDELVCHK data created.",
				"TIME_STMP" : "20210223192842"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask4Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=4)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 4,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_4",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery /ITAPC1/SAPK-001AAINITAPC1 successfully deleted",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 4,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_4",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery /ITAPC1/SAPK-001AAINITAPC1 was succesfully created",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 4,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_4",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery request SAPK+001AAINITAPC1 of type AOU was succesfully created",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 4,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_4",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery request SAPK-001AAINITAPC1 of type AOI was created",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 4,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_4",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Check Set ABAP_IN_SCP successfully assigned",
				"TIME_STMP" : "20210223192842"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask5Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=5)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 5,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_5",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Merge transport Plugin configuration loaded: Delivery transport SAPK+001AAINITAPC1",
				"TIME_STMP" : "20210223192842"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 5,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_5",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Transports merged into delivery /ITAPC1/SAPK-001AAINITAPC1",
				"TIME_STMP" : "20210223192843"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask6Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=6)/logs`,
	Body: `{
	"d" : {
			"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 6,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_6",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Release transport Plugin configuration loaded: Delivery transport SAPK+001AAINITAPC1",
				"TIME_STMP" : "20210223192843"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 6,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_6",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery Transport SAPK+001AAINITAPC1 was succesfully released",
				"TIME_STMP" : "20210223192843"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask7Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=7)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 7,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_7",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery request SAPK-001AAINITAPC1 of type AOI was merged",
				"TIME_STMP" : "20210223192844"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask8Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=8)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 8,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_8",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Required objects inserted into SAPK-001AAINITAPC1",
				"TIME_STMP" : "20210223192844"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask9Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=9)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 9,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_9",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 OLC for delivery /ITAPC1/SAPK-001AAINITAPC1 was succesfully done",
				"TIME_STMP" : "20210223192845"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask10Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=10)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 10,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_10",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Merge transport Plugin configuration loaded: Delivery transport SAPK-001AAINITAPC1",
				"TIME_STMP" : "20210223192845"
			},
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 10,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_10",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Delivery Transport SAPK-001AAINITAPC1 was succesfully released",
				"TIME_STMP" : "20210223192855"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask11Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=11)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 11,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_11",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 XML and SAR Archive /usr/sap/trans/tmp/SAPK-001AAINITAPC1.SAR successfully created and stored to file system",
				"TIME_STMP" : "20210223192855"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask12Logs = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=12)/logs`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 12,
				"log_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q_12",
				"msgty" : "I",
				"detlevel" : "3",
				"log_line" : "I:/BUILD/LOG:000 Provide log Plugin configuration loaded: Delivery /ITAPC1/SAPK-001AAINITAPC1",
				"TIME_STMP" : "20210223192855"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask0Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=0)/results`,
	Body: `{
	"d" : {
		"results" : [
		]
	}
}`,
	StatusCode: 200,
}
var buildGetTask1Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=1)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask2Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=2)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask3Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=3)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask4Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=4)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask5Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=5)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask6Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=6)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask7Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=7)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask8Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=8)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask9Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=9)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask10Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=10)/results`,
	Body: `{
		"d" : {
			"results" : [
			]
		}
	}`,
	StatusCode: 200,
}
var buildGetTask11Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=11)/results`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 11,
				"name" : "SAR_XML",
				"additional_info" : "/usr/sap/trans/tmp/SAPK-001AAINITAPC1.SAR",
				"mimetype" : "application/octet-stream"
			}
		]
	}
}`,
	StatusCode: 200,
}
var buildGetTask12Result = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=12)/results`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 12,
				"name" : "DELIVERY_LOGS.ZIP",
				"additional_info" : "SAPK-001AAINITAPC1.zip",
				"mimetype" : "application/x-zip-compressed"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask12ResultOrig = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/tasks(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=12)/results`,
	Body: `{
	"d" : {
		"results" : [
			{
				"build_id" : "AKO22FYOFYPOXHOBVKXUTX3A3Q",
				"task_id" : 12,
				"name" : "DELIVERY_LOGS.ZIP",
				"additional_info" : "504B03041400000008009B9B5752BE5CF063BB7900009036050016000000534150452D3030314141494E4954415043312E583032D49D6D6FE248D686BFAFB4FFA134A367D23DEA24C6BC84206525C7381D7A704C63279DD66A850838697608B0E07477F6D73F5506131B97AB4E958F57BB68A4E904FBAAB773EEAABBCA811A718241EDFCFCFE57D0EBAF7FA9C537B4EBC4B91F78C3800C86CEC01A5A41CFBB49DE346A245A8F179BD5721D91E57A1AAE097B75C82FBE35F8E3D8306A96D5BBE905D6C0AEFDB2BFC9249BD74D143E93CC8BDE746F986F57B549B422AB71F4EDF0AA68F576D139F91EAE37B3E5828C1753B20EE7E17813B28BEAEDDA49DD6C9FB4CE7E21BF9CB5EBFB5BCEDBE4AF7F318933B835EBC6BDF3339CBC44ECFEE52389BE8524FC19B765B50E8F57EBE524DC6C668B27F21C46DF96D30D795CAE6921FF7A09375141131B8C5CABD5585D299521C7ABD57C3619C7A56CC235AD30AD9FF1736C3CD41B8FD3FD3DF536BB67B62093F92C5C4444F0A2F7D70CE3977D3BEAF77E345E471D9269CEF2E19FE1243ADEACC2C9EC71361137EDAF7FA9C7AC86C9656DAFA26DBE35EF4797CE95377446342C46B7BE77796DF9D7492BCC06C52CA6EA776F4BAF3738772BB60310B6358384345CA6E3288CE3269A3D87DB6E350D93C667EDBCD6306BCDB37DD0D0F10C7FCE2232594EC3EC38186F81754ED4324B39159B067169ACED1AF6BF9E8374A8832F866190613DAEFF9ED23A315BE45DC2A1F79163526B9ED4CC13D360FF6C759A46C730DE9FA4182F8B593C36E162FC300FA7092B75C585DA2B7567EA9F2C627EDB474BEDFCC46095326BAC5A8D8E59EB34CF53574F968B68BD9C93C7D93CEC90D397CDFA74335E9DC6AD3D8D9E57A76C800E46E884F67F0AF137F2EB6A3625B533C338637A62D6DAE3FA63F8786C86E3C971E3C10C8F1F1A93C9B1316D371EC3D6E4F1F1BC49DE7D9F8C57EF33946DCE647E351F2F9E5EC64FE1C551CFF78E9D9BA3CCBBACCE17470775A6CD1F9F0E0F2A4C1B91BD75AB5F1754A132BF7E78797CA4A3F2BA985CBC869B6C5536CFE48258FD7E96F32D1C475B583C80E4F0B6F162B9787D9EFD7B2BAD17E488566564DFFA81E73AC36C9D323FBCD0B09A2C9F9FE31B0EFAFF8813983FC61B3219CF595C8DD91C309F2F7F6C3AFBB78F7F148CAD931F5ABD28B0978B0595405A7EB424DD4BD7A78DBDEE5E92E3E363427C6778E70C59F38F482C78F1BFD85BFE573F705CF633C51D25E9D23AA740CF757B0153AFAD82ED4B625232A1A2189179F83D9C931A59AEC2052D98CD7C63F234FB1E2EE2D820ABE56C11856B7E9EE4F389D0AA0C284E966CE90CDB898D2CCD1E5F16933802DE6E4914328D1D6FEBBDBF0618DB29C85B04EF215F1D7F77C199D9C8CFDCB4F2992C488574728191B960B698CF1621D944E368932B825DC0B2764AA5913578C3BBA0BF4BEB5D479077DBEC7E9F09A7673A796E36FB3EEB9074E6CD4F3627CF279956645373F5B27E0AA7CBC9CBDB05D93AB0E49AE5FA21DBD28809354D2BDA25C935D3D7C5F87936C984543D0ED9B3FAFD2D4D5A5AF1673AC7C4E358B8FCDA0572FC3AFD9D9C6EDF3BED5165180E9D1BFB2BEBF21A6DD3747AEC2D486F413B7B3EDF6AC8EFA73944B44EE2AB438EBA47E41D9D0E27DFDEA6DCF79C3BD8F0BD6C58938EBCA3FCFBE3F56EEEED10DA88DD6F739751955A273DC3846D9BCDB9CBE2458C20539A993BE83A86CE49AFE471BD7C268E7166EC66069AE1EFF21DFAFE2453AD7A2BBE7B166EF6F7D784F713F2FB01829879C41FAA8C3A8771A50A313890911F0C5539B91EB555BA94530DC32AD9143BF0477452E859FD92202BF046B442657B249077C82ECBDBF71BE63D92153E5DFBFF32AC0743BB6FF9F934FE859C9CA49780FEEB66BE7CEA903F89FDD9A6C9B0DE4C56E14FBA367A0A233F8C226A127E23B2577213A9A5E7DF8F614467DDD9D39AE69A4D979903AAAF1B26B41B96A51B12D803A3FB81FDAF7B19FFCFB04F8ABB24B0BABD74981537F1A067FFBE4DF8F03B9598C9875850FF41E87A801A1C8F4C1F46EB70B51CD145E7E3ECE9651D8E9EA377EF6937B05F3E47A317E69246B47F17D3F17A4A1705D1FA25FCC05CC0287A5D85F417065A590C185FB1594F18F80379FB8101FCA18D5F5614FE8CDE0ADBFDC41081731F408BEB5087F5C91B8EE88ACAA7C6671B84EC8DBE6775A9E59C3D8FE930FE19BEC2819B30228BF007194F9855DDAF56B69B09DB59302920BDF664DDC499BCD2CB26BBB78FDB37CD506578088C0081712B66E4C4C9773CBA8A1D78DDB4F27326FBD44BA677C190266620C4F1ABEC609383CAC85E65E41E989C0C1DDAC02540E71E1D19A0227BEEA03F74FAFAC8CC444281AE1358A580F93AFAB79736EAD0C440C4344D90B8A31D2371477BE8741DC4667BC17074EB3B2960BFE7DEDA83DB4B1CF06D593248E707FBBB613AAF2A437B7E5E860E93872DBA7CBC5C8C71B8711E2371E3FC6AD8736EBA259A9D47BAD6C00AC443A44EB4FBA8750CBE0E9C41196466B4BB5DC6C38E1D05F5390C6E50EAB9D47809520FC4B0AD64D55382C164B12C834EA5A519AE655729457B3EFA5A6BDF7A7C723236F8E42472A4E4A19735C83273DC3A3B8FCFF360DB06A27D709D7D889151334CC338DC8ED036FD3B1E2CBF77171FACC0D57222C3D0B4831986A61DCC302AB283BB3240B30620C0F955C6481D7E87A29351ED203F1871ECA0CAC001EDA00E5232452B2221765088D4B1838A7594DB411D20629A02EDA00E1277B4E5765009A8620775C0303B28208374BE023BC8E763D841A55C84D841550902D8414524C40E2A220176509D28B3838A44881D848F36C40EAAC78E82FA40ED60263534ED6086A16907B30C3D3B9865E8D9C12C03DF0EF2F9E86B2D543BC81F1B7C32B21DCCD92625579898387C73681A35547318F3A0D91E5F5CD21CA618DAE630C5D03687294665E6302E03348780C29D57659C44E275283A19D91CF28211CB1CC2070E6C0ED591D2095B0909338702A49E3954AA23C41CAA0311D3146C0ED591B8A30D31870A403573A80E869AC3423248E72B31873C3E8E3954C845983954932090395442C2CCA11212640E55897273A844849943E868C3CCA16AEC28A80FDC1CA65243DB1CA618DAE630CDD035876986AE394C33AA30873C3EFA5A0BD91CF2C6069F8C6E0E0F6C93A239DC9AB82ACC21EEC9A1A9727268229C1CA61825CC61F993C314A34273887A72C8AB325622557572C81B6E747205E610FDE4501D0998B0D14F0E05485D73887C72A80E444C530573887E72A8848499C38A4E0ED5C17073A87C729892898ACC615527870AB9083587E827874A48A839C43E39542542CC21FAC92174B4A1E610F9E4F00DA8620ECB9F1CA61825CC61F993C33443DF1C567B72C8E3A3AFB5D0CD61552787BCC8413487654E0EB7260EDD1C5A834F23EFF2D3E8A37383E710535060DEA7EED0B6896986AE4D4C33746D629A51954D4C958111F8DC2AA3A414B743D1C9B836911B8C48365161E0A0365103299BBAD590209B28426AD944B53A026CA20610314DA1365103893BDA009BA80254B2891A60A04D2C268374BE0A9BC8E5A3D844955C04D944450982D8443524C826AA212136519928B5896A44904D048F36C8262AC78E82FA806D623A35746D629AA16B13330C4D9B986168DAC40CA3029BC8E5A3AFB5706D22776CF0C9D83691EB9DD4BC62DAD3A11BC6C0F10356029E5B4C88C0DC4F2E2F73A29866E85AC53443D72AA6195559C5A40C8273A2C8AD324A5A713B149D8C6B15B9C188641515060E6A153590B2E95B0D09B28A22A4995F5541CCA25A2D016651038898A850B3A881C41D6F805954012A99450D30D02C1693414A5F8559E4F251CCA2A81BB5CCA2A20841CCA21A126416D59010B3A84C949A453522C82C82471B6416956347417DC066319D1ABA66F180619665E81ACE0C43D37066189A8633C3A8C07072F9E82B365CC3C91D1B7C32B6E1CCDB2F35B7B93784FA56B3EBDCD979B286C3E48324FD3470866ECFF77BDE8D701448EA2516B3805543C892DF0EFFA0B3741C148C5C41FF620FD888FE1A6DD0180C67E018096BE078B5521DBC747DB4072FEE6BD8972D642BE85D7EBA4E55AFF7C5F548E1ABFFFB7BF2EE398CC6F1571BBC178E0605F7AD9BAAD02E0ED8CC81FDCAC07D1CF2E1E8A583CB51E58AD39AD1C46243E2AF35D04872285A985CA7BD2F97CE80DEEF7EFC32F2AEBB608D5467C3F577CB062432B87751D2DA7335E34D9ED6DAE8C31ABB2858B32E496A6DF0992CA9D17AE220A9D5B8B2A4F65CCEFCC1226D74E968A5B208284BB2AB9B2E4B32A72B4DE084986DBE255914648AB01CC9D2805F846C659D2DA3EFE894215DBE1F14F247E50DB1BBFF8932647FFC01ECACE2B0725CB7EAB07206925522B0A78A1BD1BBB9AEBA11AEF441FDF2C3EDEAA5B85A19BDEA73DC1D0C2B2FE3C6C3E92B91E0CA9E7F46502A6F50794F0D86D5973194FE2954F93282FBEA0537B8AF3E7283FBFF405CDD395A8A084D8F8FDD414F67DE802CFB85EB2F94C5BEFF9114BE4A2EF611D12E0EB82959EE6B835BB2E5BE3659B2DC57E3CA96FBFEC78270D35AEB17D2E0F38E3F04A46EAEE1B9C5CCB9401D7CC822205F84C2E4E90F3FA29420D2387F28FB1B3D5047891A0150EA7C01A0CD8DE2B0C391B8BB5CBD0439A22671786817072CDBD1D00637A412A74B96499C12572A7177B28D345262A7124C876F28CA24B0345E2C4E5CBC54FCD27C95FD50A2B0250A1F491415B9BB54094325154144BB3860998A6082FB3864898AA871652A72C77BB6EB8E0EEBFEA52D2130343C01EF2CC912AA14DB52642B28C79D255971E4E020D9000E1D82667896E96B06B34C3350D12E0E587C408A0BEEE390859AA1CA156B06A3E1992B014D9860F4BE914F17F3804D681827D363525241528AFA06210F7DEB938D962CD93C4445BB38E073611EE282FB3864611EAA72C579C8689C58FBE45D6AA461314C983DECAFACBE8C3ED9CE68E87992498EB2F28D3680F88F921D5C3E5C3C47A7E923ADCAE7776F0FBCE55B1103ABD2160C3DAC1618452DF0658F74015B202AA0FA46389F2B6E84F3B97C230A661741C6E34C2E41BE5EC9ABECE482877671C06263880BEEE39065938B12573AB904FC501B0583BEDE04230042542E80CD310C976D3B688E096012CAA1432781807A33C9831FFC0240C2201A2D0C71B0BDFC9E382C6AA5E280897651C035B1032C019689837E5F88C541912B11074ACB879B8E2870419274F54734866942D9BE2D3EEF92CED969D225786F57B6244B537B25B0C2EAF62476559FEC555565C9433125C095F545891A0B8242525D700D29CABAED55D0A93158FCB7DF25C0773AE0A2898FAF4438139E79A8936FAFB2131E1EDAC5014B273C5DB07CC2D325CB263C25AE74C23339EB2B76DEAF37EB15D1802AD2BB943C23A1C6926D7D6669C5A958D84718F9D8B32A5B80EAA3457FCC5102CB79080B072CF9630EC49EC866A32257928D9486B3FCE482A479D3B358DE5CDE8E6CC9B91E605EDCB1EC6170257EF232D76BF21547524F3BFBA90C7A4B8E3DACEBC05745FC8A8AE823D5872CE0DD40870BBEB43D6417891E3F14F5FFA8980FD4DDD42F84E1C5388F5626CE73BBA39871AE565979ACCB2B5B26D64B748534D6456CA55847DB7DEE59AD5C02EE5F65E7773CB48B03E64CC49581FB3864D90CAFC495AA608B136ACE7DA0A7820530603A5903890A705085794F6132D3CE7899EE83AB4A4C977C6A5B49BAD8BECBE9C59DEC0FAB63CB26069521BC931DE50AEA59ACAD45015F6A2DC183B25301769288974A09112B9F8A78A5922A8156935425E992C02F4597A655293A60C5A53696D2EC1254572DBBF68980B37C395399FFD4962F786817077CD8157E45D87E35DCC3A58B12572ABA679C58D37C3A524053B04532BD053DDA286A58A9398A4765A989DB650911B1DBF248B5AEDB3712477FDA9AC901D01F3CB48B03961C57E88365C715FA6499062971A549D5C6DA20E580A02924D9DE8063C0BB97C5F9C7EB8F52AAC501EAEFD215C0903A5A0DA5B485A6D4E1885B45E79A4908D03A3CB48B0396ACB5D0B0FD6AB8873AA7C495A6E139EA5AAB8806CF1EBA6EC0596B1536AC946AF1A8E5D65A222262B7A9ACB5848D44D11FFB4A3339E4FAA38DCE2F895C1CB0EC301A13DCC7214B34488D2B4B2AFBEAE0083170ADAF67D697D64D2BE8377A9FDDAFB76EE3F6B3FFE98B5686C1F1C274A31CDF19DEF5ECF407F5E9920FC298B2A959EADBB7992749A070F12E2E65773DFB56832BF9BC3C0ABEB66EBA7D678851E77C8FDCF586D79E568748D8BCEFE9500D15C89E99425C6388EAB0B28F254045BB3860CE67905406EEE390C5A2AAC89588EA90F7B7ED3AF2C905097581DD32BA762C950F564E5EE24F218AC997BD9BEE887D6B4835785671C99FB114B065DF93C3E8BB1944FCF061011F2039FC51C71197AE4A74AA898B2E5AF8F0A03E56F268012AB85F4D4F1C4A8B12572A2D5D2C69E1806439D4F58762A7C36FA534F319B7A2B467E85155D51E0D83917573E3E9488ABCE67BBA962402BA8615E0DC04BDA0A7258AB00298A63BE2435188B6F2C21E435B6FFD6B7EEBD8AB9CB6EAA385DAAA8F9578615470BF9A9EC86AAB2257A2AD94566FD75B96E9D846D334ED5AAD61B41A76CD6E9F759B75A3695C6B3DC005C00A13E9D6F72EAF2D3F1DB5AABC430F766BDE8F0227F399BF8E8C29338DAC9A81421D21BCFB32BC5C278E7C17AF0F2930D06C6E91C841E2AFC4FE29C35FD94DA766D68CB386639E371B75A76674DB4DBB7E6EB60C4B3FBC2558E5F056E481C25BC2540E6F555EAED1F7A8B483E02ED783B9E086574F14DCB2E82B17DCADAB7ABBDB68D88E639F9FD7EB8ED13ABBBABA6AD58C966937F4835B86550D6E551E24B8654CD5E056E60983BB3C2D1BDC257BF030B815AA27086E69F4E90777605DDEDA77CED0675F692089E2FD75E0F5BC236D5EB6FC72ED185881511FF9D792662497A51791D6E08F63C3A85956EF66DB2050C5F70582CC42BE10122D5F26DFC2E989F8AA75380FC79BEC65CD66AB669287D78846DE8FF52C8AC245FAED603D5E6CB6F5FD1EAEBF85E329699C9DB4C9FFA52FEAD2A52B992C9F57EB7043F1B436E4FCA47970CDCB7A1CCD968B0E5D876FC2097947F5B4BD2DF794FEFCFEE4201BA2F1C39C5669B62083E38797C7C7704D36AF8BC9B7F57231FB77B609D9CB87A2CBD3FDDA341AAC6BE3FFA568E92E7C797E1EAF5F3BFC77593430D69B89E0BC1970DF6CD5765F79CB79335846E3F9FC955DE33DFC339C441BDE5566A34EBA4EDF091CD6F1CFE3C5947B5902ABB79BAC8B1289E15C99AEDEDFE39C9B86DF49B49E7C30D8EB1F845CAD67E42A7C20B573526B744CB3633489699835F99DDDD966B25C2C685B668BA7DDD174BF4F92DF2D179B8E16E40D4023E02055F9909BE50FD23B7A26D33D8C866BCCBAEE5EAAD421B92D5D05F9FD5471A2D9269A4D3664FA30A2F78ED8E0CD22F28EFE7F14B1C1BAA8B73F90F8A79FF49FEF7598EBE57CFE309EFC49DEB17FEDB0C607B2FDE9E78591A6E69BC48CE803958974773A54B7A860BDA9C13B567A2669E96DE16FD1EC39ECD0083931CC13161BE4781F2B7FFD4B8D5E3AA8350DE2524922CEFDC01B06C96F6B46AC8D0C4268301306A28DED905F1886FE775E6B98A6D1FC657F0395E29FB4E726CB69989A283A5BADDC5E737E4E7E05BDDE6EB857BBA15DDFB5830C68425B432BE87937C99B06D5E537F95C4FA926EDAA98D7E67D9D0D930AD7260A9FB3FB04F4A67BC37CBBAA4DA215598DA36F875745ABB78BCE09D5EC0D8B4DD6A53BF56717D5DBB593BAD93E699D51013C6BD7DF7AAC4D85C024CEE0D6AC1BF7CECF70F212C7361DFBE85B984C5D54E78F57EBE584AA3DCBC5E730FAB6A4FAF3B85CD342FEF5126EA282263618990D1D7D512A438E57ABF96C124F0D745E58D30AB321FC39361EEA8DC7E9FE9E7AFCEC2695F7C97C4655AC784325EE849AC18260D78EFABDCF26FA0EC93467196BEBF166154E668FB389B869F1A281B21A2697B5BD8AB6992D7B2F9D2B6FE88C68588C928575D20A26DA3493D4EFDE965E6F70EE566C07206C41A9D8ACB5FECB5291AF2BFFAB39B897D5613DAEFF9ED23A315B54D8771C7A1F93D8E6498DE92DFB67ABD3343A07D2FCB298C563132ED82A699AB052575CA8BDF8EB86FD0C10F7426E1268766A67A9ABE9AC13D119893CCEE674CA387DD9AC4F37E3D569DCDAD3E87975CA06E860844E68FFA7107F23BFAE6653523BAB370CA62766AD3DAE3F868FC766389E1C371ECCF0F8A131991C1BD376E3316C4D1E1FCF9BE4DDF7C978F53E43D9E64CE657F3F1E2E965FC145E1CF57CEFD8B939CABCCBEA7C71745067366F9E0E0F2A4C1B91BD75AB5F1754A132BFDEAE59D992F5E235DC64ABB27926176CC194E5D0B578B485C503480E6F1B2F968BD7E7D9BFB7D27A418E685546ECF921CF7586D93A657E78A161C55625F10D07FD7FC409CC1FE30D99D045268DAB319B03E6F3E58F4D67FFF6F18F82B175F243AB1705F67EF9427D47F7D2F56963E96A8E1C1F1F13C28ED39D216BFE1189052FFE177BCBFFEA078ECB7EA6B823D8290DADD6848A6244E6E1F7704E97F8CB55B860EB263AF38DC9D3EC7BB8886383AC96B34514AEF97992CF2742AB32A03859B2A5336C2736BC346BA76E797C596C97A76FB7240A99C68EB7F5DE5F038CED14E42D82F790AF8EBFBBE0CC6CE4676E5AF94C16A4423AB9C0C85C305BCC678B305EF16E7245B00B58D64EA934C69E8277417F97D6BB8E20EFB6D9FD3E134E5B1FBBEFB30E4967DEFC6473F27C9269453635572FEBA770BA9CBCBC5D90ADC36EC97FD00FD996C67696A615ED92E49AE9EB62FC3C9B64426AB79551BFBFA5496B6F7D603C8E85CBAFC49DB2D7E9EF84B3EB42DFAFD1364DA7C7DE82F4161B6620B71AF2FB690E11AD93F8EA90A3EE117947A7C3C9B7B729F73DE70E367C2F1BD6A423EF28FFFE78BD9B7B3B6C0B63F7DBDC6554A5D649CF3061DB6673EEB2781123C89446F1EE8E639C19C09D9DF8EEFAC1F922BDBF26BC9F73366DE6117FA832EA1CC6952AC4E040467E3054E5E47AD456E9524E350CAB6453ECC067CF58F5AC7E49901578235AA1B23D12C83B44BC6169F72DCEF7611C6E5DFAAF9BF9F2A943FE24F6679B26C37A3359853FE9DAE8298CFC30627B2ABF11D92BB989D4D2F3EFC730A2B3EEEC694D73CDA6CBCC01D5D70D13DA0DCBD20D09EC81D1FDC0FED7BD8CFF67D88A271CFC261EF46CC18ECD25DB68633B23EB70B564DB238FB3A79775387A8EDEBDA7DDC07EF91C8D5E984B1AD1FE5D4CC7EB295D1444EB97F0037301A3E87515D25F007677806531607CC5663D61E00FE4ED8778577068E39715853FA3B7C2763F31043BA58116D7A10EEB93371CB16D766A7CB641C8DEE87B56975ACE19DB2D257F86AF70E0268CC822FC41C6136655F7AB95ED66C276164C0A48AF3DE3CDD37C48A4974D766F1FB7166FC313C4F010180102E356CCC81FC0391E5DC50EBC2EFC8845A677C19026A6F8192F7E95A55F1DA14A0E2A237B95917B60723274680397009D7B7464808AECB983FED0117F7EBF10997DB4CFA14637B04A01F375F46F2F6DD4A1898188699A2071473B46E28EF6D0E92A1C004B81BBBFD84901FB3DF7D61EDC729E77D701DF962583747EB0BF1BA6F3AA32B4E7CB1EBCA5094E175DF0076E6138DC388F91B8717E35EC3937DD12CDCE235D6B90FDD04C0CA2DD47AD23FB639941196466B4BB5DC6C38E1D05F5390C6E50EAB9FBAF24D75FA6D956B2EA29C160B2589641A7D2D20CD7B2AB94A23D1F7DADB56F3D3E39191B7C72123952F2D0CB1A6499392E78ECA860DB40FFC1293E7064D40CD330341E6B14F260F9BDBBF86005AE96131986A61DCC3034ED608651911DDC95019A350001CEAF3246EAF03B149D8C6A07F9C188630755060E6807759092295A1109B18342A48E1D54ACA3DC0EEA0011D31468077590B8A32DB7834A40153BA80386D9410119A4F315D8413E1FC30E2AE522C40EAA4A10C00E2A222176501109B083EA44991D542442EC207CB42176503D7614D4076A0733A9A16907330C4D3B9865E8D9C12C43CF0E6619F87690CF475F6BA1DA41FED8E09391ED60CE3629B9C2C4C4E19B43D3A8A19AC39807CDF6F8E292E630C5D036872986B6394C312A33877119A0390414EEBC2AE32412AF43D1C9C8E690178C58E6103E706073A88E944ED84A4898391420F5CCA1521D21E6501D8898A66073A88EC41D6D88395400AA99437530D41C1692413A5F8939E4F171CCA1422EC2CCA19A0481CCA11212660E95902073A84A949B432522CC1C42471B660E556347417DE0E630951ADAE630C5D036876986AE394C3374CD619A518539E4F1D1D75AC8E6903736F864747378609B14CDE1D6C455610E714F0E4D95934313E1E430C528610ECB9F1CA618159A43D493435E95B112A9AA9343DE70A3932B3087E82787EA48C0848D7E722840EA9A43E493437520629A2A9843F493432524CC1C567472A80E869B43E593C3944C54640EAB3A3954C845A839443F39544242CD21F6C9A12A11620ED14F0EA1A30D3587C827876F40157358FEE430C528610ECB9F1CA619FAE6B0DA93431E1F7DAD856E0EAB3A39E4450EA2392C7372B83571E8E6D01A7C1A79979F461F9D1B3C87988202F33E7587B64D4C33746D629AA16B13D38CAA6C62AA0C8CC0E7561925A5B81D8A4EC6B589DC6044B2890A0307B5891A48D9D4AD8604D9441152CB26AAD5116013358088690AB5891A48DCD106D84415A0924DD400036D623119A4F355D8442E1FC526AAE422C8262A4A10C426AA214136510D09B189CA44A94D5423826C2278B4413651397614D4076C13D3A9A16B13D30C5D9B986168DAC40C43D326661815D8442E1F7DAD856B13B963834FC6B6895CEFA4E615D39E0EDD30068E1FB012F0DC624204E67E72799913C53443D72AA619BA5631CDA8CA2A2665109C13456E9551D28ADBA1E8645CABC80D4624ABA8307050ABA881944DDF6A4890551421CDFCAA0A6216D56A09308B1A40C444859A450D24EE7803CCA20A50C92C6A808166B1980C52FA2ACC22978F621645DDA865161545086216D59020B3A88684984565A2D42CAA114166113CDA20B3A81C3B0AEA03368BE9D4D0358B070CB32C43D77066189A8633C3D0349C1946058693CB475FB1E11A4EEED8E093B10D67DE7EA9B9CDBD21D4B79A5DE7CECE93351C261F24E9A78133747BBEDFF36E84A340522FB19805AC1A4296FC76F8079DA5E3A060E40AFA177BC046F4D76883C6603803C7485803C7AB95EAE0A5EBA33D78715F237C2576EF8BEB91C257A9AFC44645BB38603307F62B03F771C8C22FC556E58AD39AD1C46243E2AF35D04872285A985CA7BD2F97CE80DEEF7EFC32F2AEBB608D5467C3F577CB062432B87751D2DA7335E34D9ED6DA68D137DDEB63CDBA24A9B5C167B2A446EB8983A456E3CA92DA7339F3078BB4D1A5A395CA22A02CC9AE6EBA2CC99CAE34811362B6F99664519029C272244B037E11B29575B68CBEA3538674F97E50C81F9537C4EEFE27CA90FDF107B0B38AC3CA71DDAAC3CA19485689C09E2A6E44EFE6BAEA46B8D207F5CB0FB7AB97E26A65F4AACF717730ACBC8C1B0FA7AF44822B7BFE1941A9BC41E53D3518565F86EC4BD531CA08EEAB17DCE0BEFAC80DEEFF037175E7682922343D3E76073D9D7903B2EC17AEBF5016FBFE4752F82AB9D84744BB38E0A664B9AF0D6EC996FBDA64C9725F8D2B5BEEFB1F0BC24D6BAD5F4883CF3BFE1090BAB986E71633E70275F0218B807C110A93A73FFC88528248E3FCA1EC6FF4401D256A0440A9F3058036378AC30E47E2EE72F512E4889AC4E1A15D1CB06C47431BDC904A9C2E5926714A5CA9C4DDC936D248899D4A301DBEA12893C0D278B13871F152F14BF355F64389C296287C245154E4EE52250C95540411EDE280652A8209EEE390252AA2C695A9C81DEFD9AE3B3AACFB97B684C0D0F004BCB3244BA8526C4B91ADA01C779664C591838364033874089AE159A6AF19CC32CD4045BB3860F101292EB88F43166A862A57AC198C8667AE04346182D1FB463E5DCC0336A1619C4C8F4949054929EA1B843CF4AD4F365AB264F31015EDE280CF8579880BEEE3908579A8CA15E721A37162ED9377A99186C53061F6B0BFB2FA32FA643BA3A1E7492639CACA37DA00E23F4A7670F970F11C9DA68FB42A9FDFBD3DF0966F450CAC4A5B30F4B05A6014B5C0973DD2056C81A880EA1BE17CAEB811CEE7F28D28985D04198F33B904F97A25AFB2930B1EDAC5018B8D212EB88F43964D2E4A5CE9E412F0436D140CFA7A138C000851B90036C7305CB6EDA03926804928870E9D0402EACD240F7EF00B00098368B430C4C1F6F27BE2B0A8958A0326DA4501D7C40EB00458260EFA7D21160745AE441C282D1F6E3AA2C00549D2D51FD118A60965FBB6F8BC4B3A67A74997E0BD5DD9922C4DED95C00AABDB93D8557DB2575595250FC5940057D617256A2C080A4975C135A428EBB65741A7C660F1DF7E9700DFE9808B263EBE12E14C78E6A14EBEBDCA4E78786817072C9DF074C1F2094F972C9BF094B8D209CFE4ACAFD879BFDEAC574403AA48EF52F28C841A4BB6F599A515A762611F61E463CFAA6C01AA8F16FD3147092CE7212C1CB0E48F39107B229B8D8A5C4936521ACEF2930B92E64DCF627973793BB225E77A807971C7B287C195F8C9CB5CAFC9571C493DEDECA732E82D39F6B0AE035F15F12B2AA28F541FB28077031D2EF8D2F6905D247AFC50D4FFA3623E507753BF108617E33C5A9938CFED8E62C6B95A65E5B12EAF6C99582FD115D25817B195621D6DF7B967B57209B87F959DDFF1D02E0E9833115706EEE3906533BC1257AA822D4EA839F7819E0A16C080E9640D242AC04115E63D85C94C3BE365BA0FAE2A315DF2A96D25E962FB2EA71777B23FAC8E2D9B185486F04E76942BA867B1B616057CA9B5040FCA4E05D849225E2A2544AC7C2AE2954AAA045A4D5295A44B02BF145D9A56A5E8801597DA584AB34B505DB5ECDA2702CEF2E54C65FE535BBEE0A15D1CF06157F81561FBD5700F972E4A5CA9E89E71624DF3E948014DC116C9F416F468A3A861A5E6281E95A5266E972544C46ECB23D5BA6EDF481CFD696B2607407FF0D02E0E58725CA10F961D57E893651AA4C49526551B6B83940382A690647B038E01EF5E16E71FAF3F4AA91607A8BF4B570043EA683594D2169A5287236E159D6B262140EBF0D02E0E58B2D642C3F6ABE11EEA9C12579A86E7A86BAD221A3C7BE8BA0167AD55D8B052AAC5A3965B6B898888DDA6B2D6123612457FEC2BCDE490EB8F363ABF247271C0B2C3684C701F872CD12035AE2CA9ECAB8323C4C0B5BE9E595F5A37ADA0DFE87D76BFDEBA8DDBCFFEA72F5A1906C70BD38D727C6778D7B3D31FD4A74B3E0863CAA666A96FDF669E2481C2C5BBB894DDF5EC5B0DAEE4F3F228F8DABAE9F69D21469DF33D72D71B5E7B5A1D2261F3BEA7433554207B660A718D21AAC3CA3E960015EDE280399F415219B88F43168BAA225722AA43DEDFB6EBC8271724D40576CBE8DAB1543E583979893F8528265FF66EBA23F6AD21D5E059C5257FC652C0967D4F0EA3EF6610F1C387057C80E4F0471D475CBA2AD1A9262EBA68E1C383FA58C9A305A8E07E353D71282D4A5CA9B474B1A4850392E550D71F8A9D0EBF95D2CC67DC8AD29EA14755557B340C46D6CD8DA72329F29AEFE95A9208E81A56807313F4829E9628C20A609AEE880F4521DACA0B7B0C6DBDF5AFF9AD63AF72DAAA8F166AAB3E56E28551C1FD6A7A22ABAD8A5C89B6525ABD5D6F59A6631B4DD3B46BB586D16AD835BB7DD66DD68DA671ADF50017002B4CA45BDFBBBCB6FC74D4AAF20E3DD8AD793F0A9CCC67FE3A32A6CC34B26A060A7584F0EECBF0729D38F25DBC3EA4C040B3B945220789BF12FBA70C7F65379D9A5933CE1A8E79DE6CD49D9AD16D37EDFAB9D9322CFDF0966095C35B91070A6F095339BC5579B946DFA3D20E82BB5C0FE6821B5E3D5170CBA2AF5C70B7AEEAED6EA3613B8E7D7E5EAF3B46EBECEAEAAA55335AA6DDD00F6E195635B8557990E0963155835B99270CEEF2B46C7097ECC1C3E056A89E20B8A5D1A71FDC8175796BDF39439F7DA581248AF7D781D7F38EB479D9F2CBB5636005467DE45F4B9A915C965E445A833F8E0DA36659BD9B6D834015DF1708320BF94248B47C997C0BA727E2ABD6E13C1C6FB297359BAD9A491E5E231A793FD6B3280A17E9B783F578B1D9D6F77BB8FE168EA7A47176D226FF97BEA84B97AE64B27C5EADC30DC5D3DA90F393E6C1352FEB71345B2E3AA44E36E184BCABB59BF5B36DC1A7F417EF4F0ED2211A3FCC699D660B32387E78797C0CD764F3BA987C5B2F17B37F67DB90BD7C28BA3CDDB14DA3C1FA36FE5F8A96EEC397E7E7F1FAB5C37F97850363BDB908CE9B01F7CD566DF79DB79C378365349ECF5FD935DEC33FC349B4E15D6536EAA4EBF49DC0613DFF3C5E4CB99725B07ABBC9BA28D118CE95E9EAFD3D4EBA69F89D44EBC90783BDFE41C8D57A46AEC207523B27B546C76C76CC1A310DB326BFB33BDB4C968B056DCB6CF1B43B9BEEF749F2BBE562D3D182BC0168041CE42A1F72B3FC417A47CF64BA87D1788D59D7DD4B953A24B7A5AB20BF9F4A4E34DB44B3C9864C1F46F4DE111BBC5944DED1FF8F22365817F5F60712FFF493FEF3BD0E73BD9CCF1FC6933FC93BF6AF1DD6F840B63FFDBC30D2D47C9398137DA03A91EE4E870A1755AC373978C74ACF242DBD2DFC2D9A3D871D1A21278679C262831CEF63E5AF7FA9D14B07B5A6415CAA49C4B91F78C320F96DCD88C59141080D66C240B4B11DF20BC3D0FFCE6B0DB369D67ED9DF40B5F827EDB9C9721AA6668ACE562CB7D79C9F935F41AFB71BEED56E68D777ED20039AD0D6D00A7ADE4DF2A64185F94D3FD753AA49BB2AE6C5795F67C3A4C2B589C2E7EC4601BDE9DE30DFAE6A93684556E3E8DBE155D1EAEDA27342457BC3629375E94EFED945F576EDA46EB64F5A675400CFDAF5B71E6B5321308933B835EBC6BDF3339CBCC4B14DC73EFA1626731715FAE3D57A39A172CF72F1398CBE2DA9FE3C2ED7B4907FBD849BA8A0890D466643475F94CA90E3D56A3E9BC473039D18D6B4C26C087F8E8D877AE371BABFA71E3FBC49E57D329F51152BDE51893BA166B020D8B5A37EEFB399BE4332CD59C6DA7ABC598593D9E36C226E5ABC6AA0AC86C9656DAFA26D66EBDE4BE7CA1B3A231A16A364659DB4828936CD24F5BBB7A5D71B9CBB15DB01085B592A9A2635FDE7CDF3FFB254E4EBCAFF6A0EEE6575588FEBBFA7B44ECC1615F61D87DEC724B67952637ACBFED9EA348DCE8134BF2C66F1D8840BB64A9A26ACD415176A2FFEBA613F03C4BD609AE949A0B6AD56EA6A3AEB447446228FB3399D324E5F36EBD3CD78751AB7F6347A5E9DB2013A18A113DAFF29C4DFC8AFABD994987429D9607A62D6DAE3FA63F8786C86E3C971E3C10C8F1F1A93C9B1316D371EC3D6E4F1F1BC49DE7D9F8C57EF33946DCE647E351F2F9E5EC64FE1C551CFF78E9D9BA3CCBBACCE1747077566F3E6E9F0A0C2B411D95BB7FA7551CBF4C5DFC876CDCA96AC17AFE1265B95CD33B9600BA62C872EC6A32D2C1E407278DB78B15CBC3ECFFEBD95D60B7244AB32620F1079AE33CCD629F3C30B0D2BB62A896F38E8FF234E60FE186FC8842E32695C8DD91C309F2F7F6C3AFBB78F7F148CAD931F5ABD28B0F7CB176A3CBA97AE4F1B4B5773E4F8F89810769EEE0C59F38F482C78F1BFD85BFE573F705CF633C51DC18E6968B526541423320FBF8773BAC45FAEC2055B37D1996F4C9E66DFC3451C1B64B59C2DA270CDCF937C3E115A9501C5C9922D9D613BB191A5D9E3CB62BB3C7DBB2551C83476BCADF7FE1A606CA7206F11BC877C75FCDD056766233F73D385625611DE423AB9C0C85C305BCC678B305EF16E7245B00B58D64EA934C69E8277417F97D6BB8E20EFB6D9FD3E134E5B23BBEFB30E4967DEFC6473F27C9269453635572FEBA770BA9CBCBC5D90ADC36EC97FD00FD996C67696A615ED92E49AE9EB62FC3C9B64426AB79751BFBFA5496B6F7D603C8E85CBAFC49DB2D7E9EF84B3ED42DFAFD1364DA7C7DE82F4161B6620B71AF2FB690E11AD93F8EA90A3EE117947A7C3C9B7B729F73DE70E367C2F1BD6A423EF28FFFE78BD9B7B3B6C0F63F7DBDC6554A5D649CF3061DB6673EEB278115390298DF34EB355BCBDE3186706706B27BEBB7E70C048EFAF09EFE71C4E9B79C41FAA8C3A8771A50A313890911F0C5539B91EB555BA94530DC32AD9143BF0D943563DAB5F126405DE8856A86C8F04F20E11EF58DA7D8BF38518877B97FEEB66BE7CEA903F89FDD9A6C9B0DE4C56E14FBA367A0A233F8CD89ECA6F44F64A6E22B5F4FCFB318CE8AC3B7B5AD35CB3E9327340F575C38476C3B27443027B60743FB0FF752FE3FF19B6E21107BF89073DCBDF1D71A96EB14D2BD34C2647F0A6D525DBA2637B2AEB70B5641B2B8FB3A79775387A8EDEBDA71DC87EF91C8D5E98BF1AD191594CC7EB295D4E44EB97F003F30FA3E87515D25F00F685806531607CC5663D61E00FE4ED87783F7168E39715853FA3B7C2763F31043BE08116D7A1DEEC93371CB11D7A6A99B6E1CBDEE87B56979AD519DB67257F86AF70E0268CC822FC41C6136672F7EB9CED36C476FE4C0A48AF5AE36DD77C30A5175C766F1FF1166FAB14C4F010180102E356CCC89FDD391E5DFF0EBC2EFC7446A694C190A6B4F8F1307E95A5DF3AA14A0E2A237B95917B60723274680397009D7B7464808AECB983FED0117FF4BF10997D2AD0A11639B04A01F375F46F2F6DD4A1898188699A2071473B46E28EF6D0E92A9C1D4B81BB3FF64901FB3DF7D61EDC721E95D701DF962583747EB0BF1BA6F3AA32B4E7CB9ED9A5094E976BF067756138DC388F91B8717E35EC3937DD12CDCE235D6B90FDBC4D0CA2DD47AD23FB3B9B41196466B4BB5DC6C38E1D05F5390C6E50EAB9FB6F33D75FA6D956B2EA29C160B2589641A7D2D20CD7B2AB94A23D1F7DADB56F3D3E39191B7C72123952F2D0CB5A6B99AD2E7862A960C341FF992B3E7064D40CD330349E8814F260F9BDBBF86005AE96131986A61DCC3034ED608651911DDC95019A350001CEAF3246EAF03B149D8C6A07F9C188630755060E6807759092295A1109B18342A48E1D54ACA3DC0EEA0011D31468077590B8A32DB7834A40153BA80386D9410119A4F315D8413E1FC30E2AE522C40EAA4A10C00E2A222176501109B083EA44991D542442EC207CB42176503D7614D4076A0733A9A16907330C4D3B9865E8D9C12C43CF0E6619F87690CF475F6BA1DA41FED8E09391ED60CE3629B9C2C4C4E19B43D3A8A19AC39807CDF6F8E292E630C5D036872986B6394C312A33877119A0390414EEBC2AE32412AF43D1C9C8E690178C58E6103E706073A88E944ED84A4898391420F5CCA1521D21E6501D8898A66073A88EC41D6D88395400AA99437530D41C1692413A5F8939E4F171CCA1422EC2CCA19A0481CCA11212660E95902073A84A949B432522CC1C42471B660E556347417DE0E630951ADAE630C5D036876986AE394C3374CD619A518539E4F1D1D75AC8E6903736F864747378609B14CDE1D6C455610E714F0E4D95934313E1E430C528610ECB9F1CA618159A43D493435E95B112A9AA9343DE70A3932B3087E82787EA48C0848D7E722840EA9A43E493437520629A2A9843F493432524CC1C567472A80E869B43E593C3944C54640EAB3A3954C845A839443F39544242CD21F6C9A12A11620ED14F0EA1A30D3587C827876F40157358FEE430C528610ECB9F1CA619FAE6B0DA93431E1F7DAD856E0EAB3A39E4450EA2392C7372B83571E8E6D01A7C1A79979F461F9D1B3C87988202F33E7587B64D4C33746D629AA16B13D38CAA6C62AA0C8CC0E7561925A5B81D8A4EC6B589DC6044B2890A0307B5891A48D9D4AD8604D9441152CB26AAD5116013358088690AB5891A48DCD106D84415A0924DD400036D623119A4F355D8442E1FC526AAE422C8262A4A10C426AA214136510D09B189CA44A94D5423826C2278B4413651397614D4076C13D3A9A16B13D30C5D9B986168DAC40C43D326661815D8442E1F7DAD856B13B963834FC6B6895CEFA4E615D39E0EDD30068E1FB012F0DC624204E67E72799913C53443D72AA619BA5631CDA8CA2A2665109C13456E9551D28ADBA1E8645CABC80D4624ABA8307050ABA881944DDF6A4890551421CDFCAA0A6216D56A09308B1A40C444859A450D24EE7803CCA20A50C92C6A808166B1980C52FA2ACC22978F621645DDA865161545086216D59020B3A88684984565A2D42CAA114166113CDA20B3A81C3B0AEA03368BE9D4D0358B070CB32C43D77066189A8633C3D0349C1946058693CB475FB1E11A4EEED8E093B10D67DE7EA9B9CDBD21D4B79A5DE7CECE93351C261F24E9A78133747BBEDFF36E84A340522FB19805AC1A4296FC76F8079DA5E3A060E40AFA177BC046F4D76883C6603803C7485803C7AB95EAE0A5EBA33D78715F237C9B76EF8BEB91C257A96FD34645BB38603307F62B03F771C8C2EFD356E58AD39AD1C46243E22F44D04872285A985CA7BD2F97CE80DEEF7EFC32F2AEBB608D5467C3F577CB062432B87751D2DA7335E34D9ED6DAE8C31ABB2858B32E496A6DF0992CA9D17AE220A9D5B8B2A4F65CCEFCC1226D74E968A5B208284BB2AB9B2E4B32A72B4DE084986DBE255914648AB01CC9D2805F846C659D2DA3EFE894215DBE1F14F247E50DB1BBFF8932647FFC01ECACE2B0725CB7EAB07206925522B0A78A1BD1BBB9AEBA11AEF441FDF2C3EDEAA5B85A19BDEA73DC1D0C2B2FE3C6C3E92B91E0CA9E7F46502A6F50794F0D86D59721FB3E768C3282FBEA0537B8AF3E7283FBFF405CDD395A8A084D8F8FDD414F67DE802CFB85EB2F94C5BEFF9114BE4A2EF611D12E0EB82959EE6B835BB2E5BE3659B2DC57E3CA96FBFEC78270D35AEB17D2E0F38E3F04A46EAEE1B9C5CCB9401D7CC822205F84C2E4E90F3FA29420D2387F28FB1B3D5047891A0150EA7C01A0CD8DE2B0C391B8BB5CBD0439A22671786817072CDBD1D00637A412A74B96499C12572A7177B28D345262A7124C876F28CA24B0345E2C4E5CBC54FCD27C95FD50A2B0250A1F491415B9BB54094325154144BB3860998A6082FB3864898AA871652A72C77BB6EB8E0EEBFEA52D2130343C01EF2CC912AA14DB52642B28C79D255971E4E020D9000E1D82667896E96B06B34C3350D12E0E587C408A0BEEE390859AA1CA156B06A3E1992B014D9860F4BE914F17F3804D681827D363525241528AFA06210F7DEB938D962CD93C4445BB38E073611EE282FB3864611EAA72C579C8689C58FBE45D6AA461314C983DECAFACBE8C3ED9CE68E87992498EB2F28D3680F88F921D5C3E5C3C47A7E923ADCAE7776F0FBCE55B1103ABD2160C3DAC1618452DF0658F74015B202AA0FA46389F2B6E84F3B97C230A661741C6E34C2E41BE5EC9ABECE482877671C06263880BEEE39065938B12573AB904FC501B0583BEDE04230042542E80CD310C976D3B688E096012CAA1432781807A33C9831FFC0240C2201A2D0C71B0BDFC9E382C6AA5E280897651C035B1032C019689837E5F88C541912B11074ACB879B8E2870419274F54734866942D9BE2D3EEF92CED969D225786F57B6244B537B25B0C2EAF62476559FEC555565C9433125C095F545891A0B8242525D700D29CABAED55D0A93158FCB7DF25C0773AE0A2898FAF4438139E79A8936FAFB2131E1EDAC5014B273C5DB07CC2D325CB263C25AE74C23339EB2B76DEAF37EB15D1802AD2BB943C23A1C6926D7D6669C5A958D84718F9D8B32A5B80EAA3457FCC5102CB79080B072CF9630EC49EC866A32257928D9486B3FCE482A479D3B358DE5CDE8E6CC9B91E605EDCB1EC6170257EF232D76BF21547524F3BFBA90C7A4B8E3DACEBC05745FC8A8AE823D5872CE0DD40870BBEB43D6417891E3F14F5FFA8980FD4DDD42F84E1C5388F5626CE73BBA39871AE565979ACCB2B5B26D64B748534D6456CA55847DB7DEE59AD5C02EE5F65E7773CB48B03E64CC49581FB3864D90CAFC495AA608B136ACE7DA0A7820530603A5903890A705085794F6132D3CE7899EE83AB4A4C977C6A5B49BAD8BECBE9C59DEC0FAB63CB26069521BC931DE50AEA59ACAD45015F6A2DC183B25301769288974A09112B9F8A78A5922A8156935425E992C02F4597A655293A60C5A53696D2EC1254572DBBF68980B37C395399FFD4962F786817077CD8157E45D87E35DCC3A58B12572ABA679C58D37C3A524053B04532BD053DDA286A58A9398A4765A989DB650911B1DBF248B5AEDB3712477FDA9AC901D01F3CB48B03961C57E88365C715FA6499062971A549D5C6DA20E580A02924D9DE8063C0BB97C5F9C7EB8F52AAC501EAEFD215C0903A5A0DA5B485A6D4E1885B45E79A4908D03A3CB48B0396ACB5D0B0FD6AB8873AA7C495A6E139EA5AAB8806CF1EBA6EC0596B1536AC946AF1A8E5D65A222262B7A9ACB5848D44D11FFB4A3339E4FAA38DCE2F895C1CB0EC301A13DCC7214B34488D2B4B2AFBEAE0083170ADAF67D697D64D2BE8377A9FDDAFB76EE3F6B3FFE98B5686C1F1C274A31CDF19DEF5ECF407F5E9920FC298B2A959EADBB7992749A070F12E2E65773DFB56832BF9BC3C0ABEB66EBA7D678851E77C8FDCF586D79E568748D8BCEFE9500D15C89E99425C6388EAB0B28F254045BB3860CE67905406EEE390C5A2AAC89588EA90F7B7ED3AF2C905097581DD32BA762C950F564E5EE24F218AC997BD9BEE887D6B4835785671C99FB114B065DF93C3E8BB1944FCF061011F2039FC51C71197AE4A74AA898B2E5AF8F0A03E56F268012AB85F4D4F1C4A8B12572A2D5D2C69E1806439D4F58762A7C36FA534F319B7A2B467E85155D51E0D83917573E3E9488ABCE67BBA962402BA8615E0DC04BDA0A7258AB00298A63BE2435188B6F2C21E435B6FFD6B7EEBD8AB9CB6EAA385DAAA8F9578615470BF9A9EC86AAB2257A2AD94566FD75B96E9D846D334ED5AAD61B41A76CD6E9F759B75A3695C6B3DC005C00A13E9D6F72EAF2D3F1DB5AABC430F766BDE8F0227F399BF8E8C29338DAC9A81421D21BCFB32BC5C278E7C17AF0F2930D06C6E91C841E2AFC4FE29C35FD94DA766D68CB386639E371B75A76674DB4DBB7E6EB60C4B3FBC2558E5F056E481C25BC2540E6F555EAED1F7A8B483E02ED783B9E086574F14DCB2E82B17DCADAB7ABBDB68D88E639F9FD7EB8ED13ABBBABA6AD58C966937F4835B86550D6E551E24B8654CD5E056E60983BB3C2D1BDC257BF030B815AA27086E69F4E90777605DDEDA77CED0675F692089E2FD75E0F5BC236D5EB6FC72ED185881511FF9D792662497A51791D6E08F63C3A85956EF66DB2050C5F70582CC42BE10122D5F26DFC2E989F8AA75380FC79BEC65CD66ABD6240FAF118DBC1FEB5914858BF4DBC17ABCD86CEBFB3D5C7F0BC753D2383B6993FF4B5FD4A54B5732593EAFD6E186E2696DC8F949F3E09A97F5389A2D171DD2249B7042DED56A35B3BE2DF894FEE2FDC9413A44E38739ADD36C4106C70F2F8F8FE19A6C5E17936FEBE562F6EF6C1BB2970F4597A73BB6693458DFC6FF4BD1D27DF8F2FC3C5EBF76F8EFB27060AC3717C17933E0BED9AAEDBEF396F366B08CC6F3F92BBBC67BF867388936BCABCC469D749DBE1338ACE79FC78B29F7B204566F375917251AC3B9325DBDBFC749370DBF93683DF960B0D73F0871970B72153E10D324B55AA769748C26310DB326BFB33BDB4C968B056DCB6CF1B43B9BEEF749F2BBE562D3D182BC0168041CE42A1F72B3FC417A47CF64BA87D1788D59D7DD4B953A24B7A5AB20BF9F4A4E34DB44B3C9864C1F46F4DE111BBC5944DED1FF8F22365817F5F60712FFF493FEF3BD0E73BD9CCF1FC6933FC93BF6AF1DD6F840B63FFDBC30D2D47C9398137DA03A91EE4E870A1755AC373978C74ACF242DBD2DFC2D9A3D871D1A21278679C262831CEF63E5AF7FA9D14B07B5A6415CAA49C4B91F78C320F96DCD88C59141080D66C240B4B11DF20BC318A669D6E89D46F397FD0D548B7FD29E9B2CA7616AA6E86CC5727BCDF939F915F47ABBE15EED86767DD70E32A0096D0DADA0E7DD246F1A5498DFF4733DA59AB4AB625E9CF775364C2A5C9B287CCE6E14D09BEE0DF3EDAA36895664358EBE1D5E15ADDE2E3A2754B4372C365997EEE49F5D546FD74EEA66FBA4754605F0AC5D7FEBB13615029338835BB36EDC3B3FC3C94B1CDB74ECA36F61327751A13F5EAD97132AF72C179FC3E8DB92EACFE3724D0BF9D74BB8890A9AD860643674F445A90C395EADE6B3493C37D089614D2BCC86F0E7D878A8371EA7FB7BEAF1C39B54DE27F31955B1E21D95B8136A060B825D3BEAF73E9BE93B24D39C65ACADC79B5538993DCE26E2A6C5AB06CA6A985CD6F62ADA66B6EEBD74AEBCA133A261314A56D6492B9868D34C52BF7B5B7ABDC1B95BB11D80B005A4226D88D9FA2F4B45BEAEFCAFE6E05E5687F5B8FE7B4AEBC46C5161DF71E87D4C629B2735A6B7EC9FAD586DB3D2FCB298C563132ED82A699AB052575CA8BDF8EB86FD0C10F7427612303B8D7AC76CA7AEA6B34E446724F2389BD329E3F465B33EDD8C57A7716B4FA3E7D5291BA083113AA1FD9F42FC8DFCBA9A4D29FCACC1E4C4ACB5C7F5C7F0F1D80CC793E3C683191E3F3426936363DA6E3C86ADC9E3E37993BCFB3E19AFDE6720DB94C9FC6A3E5E3CBD8C9FC28BA39EEF1D3B3747997759952F8E0EAACCA6CDD3E1417D691BB2B76EE5EB820A54E6D7DB252B5BB15EBC869B6C5536CFE482AD97B21CBA168FB6B078FCC8E16DE3C572F1FA3CFBF756592FC811ADCA883D3FE4B9CE305BA7CC0F2F34AAD8A224BEE1A0FB8F3871F963BC2113BAC6A461356653C07CBEFCB1E9ECDF3EFE5130B44E7E64F582C0DEAF5EA8EFE85EBA3E6D2C5DCC91E3E36342D871BA3364CD3F22B1DEC5FF626FF95FFDC071D9CF1477043BA5A1D59A504D8CC83CFC1ECEE90A7FB90A176CD94427BE31799A7D0F17716C90D572B688C2353F4DF2E9446855061427CBB57482EDB44696658F2F8BEDEAF4ED964420D3D8F1B6DEFB6B80B19D82BC45F01EF2D5F177179C998DFCC4DD21D92C488574728191B960B698CF1661BCE0DDE48A6017B0AC9D52658C2D05EF82FE2EAD771D41DE6DB3FB7D269CB63E76DF671D92CEBCF9C9E6E4F924D38A6C6AAE5ED64FE174397979BB205B87DD8AFFA01FB22D8DDD2C4D2BDA25C935D3D7C5F87936C984D46E2BA37E7F4B93D6DEDAC0781C0B575F893965AFD3DF0967D785BE5FA36D9A4E8FBD05E92D36CC3F6E35E4F7D31C225A27F1D52147DD23F28ECE86936F6F33EE7BCE1D6CF85E36AC4947DE51FEFDF17A37F576D816C6EEB7B9CBA84AAD939E61C2B6CDE6DC65F11A4690298DE2DD1DC73833803B3BF1DDF583F3457A7F4D783FE76CDACC23FE5065D4398C2B5588C1818CFC60A8CAC9F5A8ADD2A59C6A1856C9A6D881CF9EB1EA59FD92202BF046B442657B2490778878C3D2EE5B9CEFC338DCBAF45F37F3E55387FC49ECCF364D86F566B20A7FD2B5D15318F961C4B6547E23B2577213CD9E14FA6318D15977F6B4A6B966D355E680EAEB8609ED8665E98604F6C0E87E60FFEB5EC6FF336CC5130E7E130F7A56BE67B54DF973E89ED525DBA1635B2AEB70B564FB2A8FB3A79775387A8EDEBDA71DC87EF91C8D5E98BD1AD191594CC7EB295D4E44EB97F003B30FA3E87515D25F00B685806531607CC5663D61E00FE4ED87783B7168E39715853FA3B7C2763F31043BDF8116D7A1D6EC93371CB10D7AEA98B6E1CBDEE87B56977AD519DB66257F86AF70E0268CC822FC41C613E671F7EB9CED2EC476FE4C0A48AF5AE35DD77C30A5175C766F1FF1166FA714C4F010180102E356CCC81FDD391E5DFF0EBC2EFC7046A694C190A6B4F8E9307E95A55F3AA14A0E2A237B95917B60723274680397009D7B7464808AECB983FED0117FF2BF10997D28D0A11639B04A01F375F46F2F6DD4A1898188699A2071473B46E28EF6D0E92A1C1D4B81BBBFF54901FB3DF7D61EDC729E94D701DF962583747EB0BF1BA6F3AA32B4E7CB1ED9A5094E976BF047756138DC388F91B8717E35EC3937DD12CDCE235D6B90FDB84D0CA2DD47AD23FB339B41196466B4BB5DC6C38E1D05F5390C6E50EAB9FB2F33D75FA6D956B2EA29C160B2589641A7D2D20CD7B2AB94A23D1F7DADB56F3D3E39191B7C72123952F2D0CB5A6B99AD2E7860A960C341FF912B3E7064D40CD330341E8814F260F9BDBBF86005AE96131986A61DCC3034ED608651911DDC95019A350001CEAF3246EAF03B149D8C6A07F9C188630755060E6807759092295A1109B18342A48E1D54ACA3DC0EEA0011D31468077590B8A32DB7834A40153BA80386D9410119A4F315D8413E1FC30E2AE522C40EAA4A10C00E2A222176501109B083EA44991D542442EC207CB42176503D7614D4076A0733A9A16907330C4D3B9865E8D9C12C43CF0E6619F87690CF475F6BA1DA41FED8E09391ED60CE3629B9C2C4C4E19B43D3A8A19AC39807CDF6F8E292E630C5D036872986B6394C312A33877119A0390414EEBC2AE32412AF43D1C9C8E690178C58E6103E706073A88E944ED84A4898391420F5CCA1521D21E6501D8898A66073A88EC41D6D88395400AA99437530D41C1692413A5F8939E4F171CCA1422EC2CCA19A0481CCA11212660E95902073A84A949B432522CC1C42471B660E556347417DE0E630951ADAE630C5D036876986AE394C3374CD619A518539E4F1D1D75AC8E6903736F864747378609B14CDE1D6C455610E714F0E4D95934313E1E430C528610ECB9F1CA618159A43D493435E95B112A9AA9343DE70A3932B3087E82787EA48C0848D7E722840EA9A43E493437520629A2A9843F493432524CC1C567472A80E869B43E593C3944C54640EAB3A3954C845A839443F39544242CD21F6C9A12A11620ED14F0EA1A30D3587C827876F40157358FEE430C528610ECB9F1CA619FAE6B0DA93431E1F7DAD856E0EAB3A39E4450EA2392C7372B83571E8E6D01A7C1A79979F461F9D1B3C87988202F33E7587B64D4C33746D629AA16B13D38CAA6C62AA0C8CC0E7561925A5B81D8A4EC6B589DC6044B2890A0307B5891A48D9D4AD8604D9441152CB26AAD5116013358088690AB5891A48DCD106D84415A0924DD400036D623119A4F355D8442E1FC526AAE422C8262A4A10C426AA214136510D09B189CA44A94D5423826C2278B4413651397614D4076C13D3A9A16B13D30C5D9B986168DAC40C43D326661815D8442E1F7DAD856B13B963834FC6B6895CEFA4E615D39E0EDD30068E1FB012F0DC624204E67E72799913C53443D72AA619BA5631CDA8CA2A2665109C13456E9551D28ADBA1E8645CABC80D4624ABA8307050ABA881944DDF6A4890551421CDFCAA0A6216D56A09308B1A40C444859A450D24EE7803CCA20A50C92C6A808166B1980C52FA2ACC22978F621645DDA865161545086216D59020B3A88684984565A2D42CAA114166113CDA20B3A81C3B0AEA03368BE9D4D0358B070CB32C43D77066189A8633C3D0349C1946058693CB475FB1E11A4EEED8E093B10D67DE7EA9B9CDBD21D4B79A5DE7CECE93351C261F24E9A78133747BBEDFF36E84A340522FB19805AC1A4296FC76F8079DA5E3A060E40AFA177BC046F4D76883C6603803C7485803C7AB95EAE0A5EBA33D78715F237C9976EF8BEB91C257A92FD34645BB38603307F62B03F771C8C2AFD356E58AD39AD1C46243E22F44D04872285A985CA7BD2F97CE80DEEF7EFC32F2AEBB608D5467C3F577CB062432B87751D2DA7335E34D9ED6DAE8C31ABB2858B32E496A6DF0992CA9D17AE220A9D5B8B2A4F65CCEFCC1226D74E968A5B208284BB2AB9B2E4B32A72B4DE084986DBE255914648AB01CC9D2805F846C659D2DA3EFE894215DBE1F14F247E50DB1BBFF8932647FFC01ECACE2B0725CB7EAB07206925522B0A78A1BD1BBB9AEBA11AEF441FDF2C3EDEAA5B85A19BDEA73DC1D0C2B2FE3C6C3E92B91E0CA9E7F46502A6F50794F0D86D59721FB3A768C3282FBEA0537B8AF3E7283FBFF405CDD395A8A084D8F8FDD414F67DE802CFB85EB2F94C5BEFF9114BE4A2EF611D12E0EB82959EE6B835BB2E5BE3659B2DC57E3CA96FBFEC78270D35AEB17D2E0F38E3F04A46EAEE1B9C5CCB9401D7CC822205F84C2E4E90F3FA29420D2387F28FB1B3D5047891A0150EA7C01A0CD8DE2B0C391B8BB5CBD0439A22671786817072CDBD1D00637A412A74B96499C12572A7177B28D345262A7124C876F28CA24B0345E2C4E5CBC54FCD27C95FD50A2B0250A1F491415B9BB54094325154144BB3860998A6082FB3864898AA871652A72C77BB6EB8E0EEBFEA52D2130343C01EF2CC912AA14DB52642B28C79D255971E4E020D9000E1D82667896E96B06B34C3350D12E0E587C408A0BEEE390859AA1CA156B06A3E1992B014D9860F4BE914F17F3804D681827D363525241528AFA06210F7DEB938D962CD93C4445BB38E073611EE282FB3864611EAA72C579C8689C58FBE45D6AA461314C983DECAFACBE8C3ED9CE68E87992498EB2F28D3680F88F921D5C3E5C3C47A7E923ADCAE7776F0FBCE55B1103ABD2160C3DAC1618452DF0658F74015B202AA0FA46389F2B6E84F3B97C230A661741C6E34C2E41BE5EC9ABECE482877671C06263880BEEE39065938B12573AB904FC501B0583BEDE04230042542E80CD310C976D3B688E096012CAA1432781807A33C9831FFC0240C2201A2D0C71B0BDFC9E382C6AA5E280897651C035B1032C019689837E5F88C541912B11074ACB879B8E2870419274F54734866942D9BE2D3EEF92CED969D225786F57B6244B537B25B0C2EAF62476559FEC555565C9433125C095F545891A0B8242525D700D29CABAED55D0A93158FCB7DF25C0773AE0A2898FAF4438139E79A8936FAFB2131E1EDAC5014B273C5DB07CC2D325CB263C25AE74C23339EB2B76DEAF37EB15D1802AD2BB943C23A1C6926D7D6669C5A958D84718F9D8B32A5B80EAA3457FCC5102CB79080B072CF9630EC49EC866A32257928D9486B3FCE482A479D3B358DE5CDE8E6CC9B91E605EDCB1EC6170257EF232D76BF21547524F3BFBA90C7A4B8E3DACEBC05745FC8A8AE823D5872CE0DD40870BBEB43D6417891E3F14F5FFA8980FD4DDD42F84E1C5388F5626CE73BBA39871AE565979ACCB2B5B26D64B748534D6456CA55847DB7DEE59AD5C02EE5F65E7773CB48B03E64CC49581FB3864D90CAFC495AA608B136ACE7DA0A7820530603A5903890A705085794F6132D3CE7899EE83AB4A4C977C6A5B49BAD8BECBE9C59DEC0FAB63CB26069521BC931DE50AEA59ACAD45015F6A2DC183B25301769288974A09112B9F8A78A5922A8156935425E992C02F4597A655293A60C5A53696D2EC1254572DBBF68980B37C395399FFD4962F786817077CD8157E45D87E35DCC3A58B12572ABA679C58D37C3A524053B04532BD053DDA286A58A9398A4765A989DB650911B1DBF248B5AEDB3712477FDA9AC901D01F3CB48B03961C57E88365C715FA6499062971A549D5C6DA20E580A02924D9DE8063C0BB97C5F9C7EB8F52AAC501EAEFD215C0903A5A0DA5B485A6D4E1885B45E79A4908D03A3CB48B0396ACB5D0B0FD6AB8873AA7C495A6E139EA5AAB8806CF1EBA6EC0596B1536AC946AF1A8E5D65A222262B7A9ACB5848D44D11FFB4A3339E4FAA38DCE2F895C1CB0EC301A13DCC7214B34488D2B4B2AFBEAE0083170ADAF67D697D64D2BE8377A9FDDAFB76EE3F6B3FFE98B5686C1F1C274A31CDF19DEF5ECF407F5E9920FC298B2A959EADBB7992749A070F12E2E65773DFB56832BF9BC3C0ABEB66EBA7D678851E77C8FDCF586D79E568748D8BCEFE9500D15C89E99425C6388EAB0B28F254045BB3860CE67905406EEE390C5A2AAC89588EA90F7B7ED3AF2C905097581DD32BA762C950F564E5EE24F218AC997BD9BEE887D6B4835785671C99FB114B065DF93C3E8BB1944FCF061011F2039FC51C71197AE4A74AA898B2E5AF8F0A03E56F268012AB85F4D4F1C4A8B12572A2D5D2C69E1806439D4F58762A7C36FA534F319B7A2B467E85155D51E0D83917573E3E9488ABCE67BBA962402BA8615E0DC04BDA0A7258AB00298A63BE2435188B6F2C21E435B6FFD6B7EEBD8AB9CB6EAA385DAAA8F9578615470BF9A9EC86AAB2257A2AD94566FD75B96E9D846D334ED5AAD61B41A76CD6E9F759B75A3695C6B3DC005C00A13E9D6F72EAF2D3F1DB5AABC430F766BDE8F0227F399BF8E8C29338DAC9A81421D21BCFB32BC5C278E7C17AF0F2930D06C6E91C841E2AFC4FE29C35FD94DA766D68CB386639E371B75A76674DB4DBB7E6EB60C4B3FBC2558E5F056E481C25BC2540E6F555EAED1F7A8B483E02ED783B9E086574F14DCB2E82B17DCADAB7ABBDB68D88E639F9FD7EB8ED13ABBBABA6AD58C966937F4835B86550D6E551E24B8654CD5E056E60983BB3C2D1BDC257BF030B815AA27086E69F4E90777605DDEDA77CED0675F692089E2FD75E0F5BC236D5EB6FC72ED185881511FF9D792662497A51791D6E08F63C3A85956EF66DB2050C5F70582CC42BE10122D5F26DFC2E989F8AA75380FC79BEC65CD66ABD6200FAF118DBC1FEB5914858BF4DBC17ABCD86CEBFB3D5C7F0BC753D2383B6993FF4B5FD4A54B5732593EAFD6E186E2696DC8F949F3E09A97F5389A2D171DD2209B7042DED5EAE7467D5BF029FDC5FB93837488C60F735AA7D9820C8E1F5E1E1FC335D9BC2E26DFD6CBC5ECDFD936642F1F8A2E4F776CD368B0BE8DFF97A2A5FBF0E5F979BC7EEDF0DF65E1C0586F2E82F366C07DB355DB7DE72DE7CD60198DE7F357768DF7F0CF70126D7857998D3AE93A7D277058CF3F8F1753EE6509ACDE6EB22E4A34867365BA7A7F8F936E1A7E27D17AF2C160AF7F10E22E17E42A7C20A6496A66A751EFD44D621A664D7E6777B6992C170BDA96D9E2697736DDEF93E477CBC5A6A3057903D00838C8553EE466F983F48E9EC9740FA3F11AB3AEBB972A75486E4B57417E3F959C68B68966930D993E8CE8BD233678B388BCA3FF1F456CB02EEAED0F24FEE927FDE77B1DE67A399F3F8C277F9277EC5F3BACF1816C7FFA7961A4A9F9263127FA407522DD9D0E152EAA586F72F08E959E495A7A5BF85B347B0E3B34424E0CF384C50639DEC7CA5FFF52A3970E6A4D83B8549388733FF08641F2DB9A118B2383101ACC848168633BE41786314CD3ACD190AF9BBFEC6FA05AFC93F6DC64390D533345672B96DB6BCECFC9AFA0D7DB0DF76A37B4EBBB7690014D686B68053DEF2679D3A0C2FCA69FEB29D5A45D15F3E2BCAFB36152E1DA44E17376A380DE746FBCB5DE6893684556E3E8DBE155D1EAEDA27342457BC3629375E94EFED945F576EDA46EB64F5A675400CFDAF5B71E6B5321308933B835EBC6BDF3339CBCC4B14DC73EFA1626731715FAE3D57A39A172CF72F1398CBE2DA9FE3C2ED7B4907FBD849BA8A0890D466643475F94CA90E3D56A3E9BC473039D18D6B4C26C087F8E8D877AE371BABFA71E3FBC49E57D329F51152BDE51893BA166B020D8B5A37EEFB399BE4332CD59C6DA7ABC598593D9E36C226E5ABC6AA0AC86C9656DAFA26D66EBDE4BE7CA1B3A231A16A364659DB4828936CD24F5BBB7A5D71B9CBB15DB01085B402AB68CBAD1FE2F4B45BEAEFCAFE6E05E5687F5B8FE7B4AEBC46C5161DF71E87D4C629B2735A6B7EC9FAD4ED3E81C48F3CB62168F4DB860ABA469C24A5D71A1F6E2AF1BF63340DC0BD949A0D531EA1DE33C75359D75223A2391C7D99C4E19A72F9BF5E966BC3A8D5B7B1A3DAF4ED9001D8CD009EDFF14E26FE4D7D56C4A1721CD5A8BE989596B8FEB8FE1E3B1198E27C78D07333C7E684C26C7C6B4DD780C5B93C7C7F32679F77D325EBDCF50B63993F9D57CBC787A193F8517473DDF3B766E8E32EFB23A5F1C1DD499CD9BA7C3830AD346646FDDEAD70555A8CCAFB76B56B664BD780D37D9AA6C9EC9055B30653974311E6D61F10092C3DBC68BE5E2F579F6EFADB45E90235A95117B80C8739D61B64E991F5E6858B155497CC341FF1F7102F3C7784326749149E36ACCE680F97CF963D3D9BF7DFCA3606C9DFCD0EA4581BD5FBE50E3D1BD747DDA58BA9A23C7C7C784B0F37467C89A7F4462C18BFFC5DEF2BFFA81E3B29F29EE08764C43AB35A1A2189179F83D9CD325FE72152ED8BA89CE7C63F234FB1E2EE2D820ABE56C11856B7E9EE4F389D0AA0C284E966CE90CDB898D2CCD1E5F16DBE5E9DB2D8942A6B1E36DBDF7D700633B05798BE03DE4ABE3EF2E38331BF999BB43B259900AE9E4022373C16C319F2DC278C5BBC915C12E60593BA5D2187B0ADE05FD5D5AEF3A82BCDB66F7FB4C386D8DECBECF3A249D79F393CDC9F349A615D9D45CBDAC9FC2E972F2F27641B60EBB25FF413F645B1ADB599A56B44B926BA6AF8BF1F36C9209A9DD5E46FDFE9626ADBDF581F138162EBF1277CA5EA7BF13CEB60B7DBF46DB349D1E7B0BD25B6C9881DC6AC8EFA73944B44EE2AB438EBA47E41D9D0E27DFDEA6DCF79C3BD8F0BD6C58938EBCA3FCFBE3F56EEEEDB03D8CDD6F739751955A273DC3846D9BCDB9CBE2458C20539AC5DB3B8E716600B776E2BBEB07078CF4FE9AF07ECEE1B49947FCA1CAA8731857AA10830319F9C1509593EB515BA54B39D530AC924DB1039F3D64D5B3FA254156E08D6885CAF64820EF10F18EA5DDB7385F8871B877E9BF6EE6CBA70EF993D89F6D9A0CEBCD6415FEA46BA3A730F2C388EDA9FC4664AFE426EAE953E88F614467DDD9D39AE69A4D979903AAAF1B26B41B96A51B12D803A3FB81FDAF7B19FFCFB0158F38F84D3CE859F9A6559CF2D4DD0037AD2ED9161DDB535987AB25DB58799C3DBDACC3D173F4EE3DED40F6CBE768F4C2FCD5888ECC623A5E4FE972225ABF841F987F1845AFAB90FE02B02F042C8B01E32B36EB09037F206F3FC4FB89431BBFAC28FC19BD15B6FB8921D8010FB4B80EF5669FBCE188EDD053CBB40D5FF646DFB3BAD4ACCED83E2BF9337C850337614416E10F329E3093BB5FE76CB721B6F36752407AD51A6FBBE68329BDE0B27BFB88B7785BA5208687C0081018B76246FEECCEF1E8FA77E075E1A73332A50C8634A5C58F87F1AB2CFDD60955725019D9AB8CDC039393A1431BB804E8DCA323035464CF1DF4878EF8A3FF85C8EC53810EB5C881550A98AFA37F7B69A30E4D0C444CD304893BDA311277B4874E57E1EC580ADCFDB14F0AD8EFB9B7F6E096F3A8BC0EF8B62C19A4F383FDDD309D5795A13D5FF6CC2E4D70BA5C833FAB0BC3E1C6798CC48DF3AB61CFB9E99668761EE95A83ECE76D6210ED3E6A1DD9DFD90CCA2033A3DDED321E76EC28A8CF61708352CFDD7F9BB9FE32CDB692554F090693C5B20C3A959666B8965DA514EDF9E86BAD7DEBF1C9C9D8E09393C89192875ED65ACB6C75C1134B051B0EFACF5CF18123A3669886A1F144A49007CBEFDDC5072B70B59CC83034ED6086A16907338C8AECE0AE0CD0AC0108707E95315287DFA1E864543BC80F461C3BA83270403BA883944CD18A48881D142275ECA0621DE57650078898A6403BA883C41D6DB91D5402AAD8411D30CC0E0AC8209DAFC00EF2F9187650291721765055820076501109B1838A48801D5427CAECA022116207E1A30DB183EAB1A3A03E503B98490D4D3B986168DAC12C43CF0E66197A7630CBC0B7837C3EFA5A0BD50EF2C7069F8C6C0773B649C91526260EDF1C9A460DD51CC63C68B6C7179734872986B6394C31B4CD618A5199398CCB00CD21A070E755192791781D8A4E463687BC60C43287F081039B4375A474C25642C2CCA100A9670E95EA083187EA40C434059B437524EE6843CCA10250CD1CAA83A1E6B0900CD2F94ACC218F8F630E157211660ED52408640E95903073A88404994355A2DC1C2A1161E6103ADA3073A81A3B0AEA033787A9D4D036872986B6394C3374CD619AA16B0ED38C2ACC218F8FBED6423687BCB1C127A39BC303DBA4680EB726AE0A73887B7268AA9C1C9A0827872946097358FEE430C5A8D01CA29E1CF2AA8C9548559D1CF2861B9D5C8139443F39544702266CF493430152D71C229F1CAA0311D354C11CA29F1C2A2161E6B0A293437530DC1C2A9F1CA664A2227358D5C9A1422E42CD21FAC9A112126A0EB14F0E55891073887E72081D6DA839443E397C03AA98C3F227872946097358FEE430CDD03787D59E1CF2F8E86B2D747358D5C9212F7210CD619993C3AD89433787D6E0D3C8BBFC34FAE8DCE039C4141498F7A93BB46D629AA16B13D30C5D9B9866546513536560043EB7CA2829C5ED507432AE4DE40623924D541838A84DD440CAA66E3524C8268A905A3651AD8E009BA801444C53A84DD440E28E36C026AA00956CA2061868138BC9209DAFC22672F9283651251741365151822036510D09B2896A48884D54264A6DA21A116413C1A30DB289CAB1A3A03E609B984E0D5D9B9866E8DAC40C43D32666189A3631C3A8C02672F9E86B2D5C9BC81D1B7C32B64DE47A2735AF98F674E8863170FC809580E716132230F793CBCB9C28A619BA5631CDD0B58A69465556312983E09C2872AB8C9256DC0E4527E35A456E302259458581835A450DA46CFA564382ACA20869E6575510B3A8564B8059D40022262AD42C6A2071C71B601655804A6651030C348BC56490D2576116B97C14B328EA462DB3A8284210B3A8860499453524C42C2A13A566518D08328BE0D1069945E5D851501FB0594CA786AE593C60986519BA8633C3D0349C1986A6E1CC302A309C5C3EFA8A0DD77072C7069F8C6D38F3F64BCD6DEE0DA1BED5EC3A77769EACE130F920493F0D9CA1DBF3FD9E77231C05927A89C52C60D510B2E4B7C33FE82C1D07052357D0BFD80336A2BF461B3406C3193846C21A385EAD54072F5D1FEDC18BFB1AE1DBB47B5F5C8F14BE4A7D9B362ADAC5019B39B05F19B88F43167E9FB62A579CD68C26161B127F2182469243D1C2E43AED7DB97406F47EF7E3979177DD056BA43A1BAEBF5B362091C1BD8B92D69EAB196FF2B4D6461FD6D845C19A7549526B83CF64498DD6130749ADC69525B5E772E60F1669A34B472B95454059925DDD745992395D690227C46CF32DC9A2205384E5489606FC22642BEB6C197D47A70CE9F2FDA0903F2A6F88DDFD4F9421FBE30F6067158795E3BA5587953390AC12813D55DC88DECD75D58D70A50FEA971F6E572FC5D5CAE8559FE3EE60587919371E4E5F890457F6FC3382527983CA7B6A30ACBE0CD9F7B1639411DC572FB8C17DF5911BDCFF07E2EACED15244687A7CEC0E7A3AF30664D92F5C7FA12CF6FD8FA4F05572B18F887671C04DC9725F1BDC922DF7B5C992E5BE1A57B6DCF73F16849BD65ABF90069F77FC212075730DCF2D66CE05EAE0431601F92214264F7FF811A50491C6F943D9DFE8813A4AD4088052E70B006D6E14871D8EC4DDE5EA25C8113589C343BB3860D98E8636B82195385DB24CE294B85289BB936DA491123B95603A7C43512681A5F16271E2E2A5E297E6ABEC8712852D51F848A2A8C8DDA54A182AA90822DAC501CB540413DCC7214B54448D2B53913BDEB35D777458F72F6D0981A1E1097867499650A5D896225B4139EE2CC98A230707C90670E81034C3B34C5F3398659A818A7671C0E203525C701F872CD40C55AE5833180DCF5C0968C204A3F78D7CBA98076C42C338991E93920A9252D4370879E85B9F6CB464C9E6212ADAC5019F0BF31017DCC7210BF350952BCE4346E3C4DA27EF52230D8B61C2EC617F65F565F4C9764643CF934C7294956FB401C47F94ECE0F2E1E2393A4D1F69553EBF7B7BE02DDF8A185895B660E861B5C0286A812F7BA40BD8025101D537C2F95C71239CCFE51B5130BB08321E677209F2F54A5E6527173CB48B03161B435C701F872C9B5C94B8D2C925E087DA2818F4F526180110A272016C8E61B86CDB41734C0093500E1D3A0904D49B491EFCE017001206D168618883EDE5F7C461512B15074CB48B02AE891D6009B04C1CF4FB422C0E8A5C8938505A3EDC7444810B92A4AB3FA2314C13CAF66DF1799774CE4E932EC17BBBB225599ADA2B811556B727B1ABFA64AFAA2A4B1E8A2901AEAC2F4AD458101492EA826B4851D66DAF824E8DC1E2BFFD2E01BED301174D7C7C25C299F0CC439D7C7B959DF0F0D02E0E583AE1E982E5139E2E5936E12971A5139EC9595FB1F37EBD59AF88065491DEA5E4190935966CEB334B2B4EC5C23EC2C8C79E55D902541F2DFA638E1258CE43583860C91F7320F644361B15B9926CA4349CE5271724CD9B9EC5F2E6F276644BCEF500F3E28E650F832BF19397B95E93AF38927ADAD94F65D05B72EC615D07BE2AE25754441FA93E6401EF063A5CF0A5ED21BB48F4F8A1A8FF47C57CA0EEA67E210C2FC679B432719EDB1DC58C73B5CACA635D5ED932B15EA22BA4B12E622BC53ADAEE73CF6AE51270FF2A3BBFE3A15D1C306722AE0CDCC721CB667825AE54055B9C5073EE033D152C8001D3C91A485480832ACC7B0A939976C6CB741F5C5562BAE453DB4AD2C5F65D4E2FEE647F581D5B3631A80CE19DEC285750CF626D2D0AF8526B091E949D0AB09344BC544A8858F954C42B955409B49AA42A4997047E29BA34AD4AD1012B2EB5B1946697A0BA6AD9B54F049CE5CB99CAFCA7B67CC143BB38E0C3AEF02BC2F6ABE11E2E5D94B852D13DE3C49AE6D391029A822D92E92DE8D14651C34ACD513C2A4B4DDC2E4B8888DD9647AA75DDBE9138FAD3D64C0E80FEE0A15D1CB0E4B8421F2C3BAED027CB3448892B4DAA36D6062907044D21C9F6061C03DEBD2CCE3F5E7F94522D0E507F97AE0086D4D16A28A52D34A50E47DC2A3AD74C4280D6E1A15D1CB064AD8586ED57C33DD43925AE340DCF51D75A453478F6D07503CE5AABB061A5548B472DB7D6121111BB4D65AD256C248AFED8579AC921D71F6D747E49E4E2806587D198E03E0E59A2416A5C5952D957074788816B7D3DB3BEB46E5A41BFD1FBEC7EBD751BB79FFD4F5FB4320C8E17A61BE5F8CEF0AE67A73FA84F977C10C6944DCD52DFBECD3C4902858B777129BBEBD9B71A5CC9E7E551F0B575D3ED3B438C3AE77BE4AE37BCF6B43A44C2E67D4F876AA840F6CC14E21A435487957D2C012ADAC501733E83A432701F872C165545AE445487BCBF6DD7914F2E48A80BEC96D1B563A97CB072F2127F0A514CBEECDD7447EC5B43AAC1B38A4BFE8CA5802DFB9E1C46DFCD20E2870F0BF800C9E18F3A8EB87455A2534D5C74D1C28707F5B192470B50C1FD6A7AE2505A94B85269E962490B0724CBA1AE3F143B1D7E2BA599CFB815A53D438FAAAAF668188CAC9B1B4F4752E435DFD3B52411D035AC00E726E8053D2D518415C034DD111F8A42B49517F618DA7AEB5FF35BC75EE5B4551F2DD4567DACC40BA382FBD5F444565B15B9126DA5B47ABBDEB24CC7369AA669D76A0DA3D5B06B76FBACDBAC1B4DE35AEB012E00569848B7BE77796DF9E9A855E51D7AB05BF37E143899CFFC75644C996964D50C14EA08E1DD97E1E53A71E4BB787D48818166738B440E127F25F64F19FECA6E3A35B3669C351CF3BCD9A83B35A3DB6EDAF573B36558FAE12DC12A87B7220F14DE12A67278ABF2728DBE47A51D0477B91ECC0537BC7AA2E096455FB9E06E5DD5DBDD46C3761CFBFCBC5E778CD6D9D5D555AB66B44CBBA11FDC32AC6A70ABF220C12D63AA06B7324F18DCE569D9E02ED98387C1AD503D41704BA34F3FB803EBF2D6BE73863EFB4A034914EFAF03AFE71D69F3B2E5976BC7C00A8CFAC8BF963423B92CBD88B4067F1C1B46CDB27A37DB06812ABE2F106416F2859068F932F9164E4FC457ADC37938DE642F6B365BB50679788D68E4FD58CFA2285CA4DF0ED6E3C5665BDFEFE1FA5B389E92C6D9499BFC5FFAA22E5DBA92C9F279B50E37144F6B43CE4F9A07D7BCACC7D16CB9E89006D98413F2AE563F37EADB824FE92FDE9F1CA443347E98D33ACD166470FCF0F2F818AEC9E67531F9B65E2E66FFCEB6217BF9507479BA639B4683F56DFCBF142DDD872FCFCFE3F56B87FF2E0B07C67A73119C3703EE9BADDAEE3B6F396F06CB683C9FBFB26BBC877F869368C3BBCA6CD449D7E93B81C37AFE79BC98722F4B60F576937551A2319C2BD3D5FB7B9C74D3F03B89D6930F067BFD831077B92057E103314D526B758C7AA75627A661D6E47776679BC972B1A06D992D9E7667D3FD3E497EB75C6C3A5A9037008D80835CE5436E963F48EFE8994CF7301AAF31EBBA7BA95287E4B67415E4F753C989669B6836D990E9C388DE3B6283378BC83BFAFF51C406EBA2DEFE40E29F7ED27FBED761AE97F3F9C378F22779C7FEB5C31A1FC8F6A79F17469A9A6F1273A20F5427D2DDE950E1A28AF52607EF58E999A4A5B785BF45B3E7B04323E4C4304F586C90E37DACFCF52F357AE9A0D634884B358938F7036F1824BFAD19B1383208A1C14C188836B6437E6118C334CD5ACBA8D7EABFEC6FA05AFC93F6DC64390D533345672B96DB6BCECFC9AFA0D7DB0DF76A37B4EBBB7690014D686B68053DEF2679D3A0C2FCA69FEB29D5A45D15F3E2BCAFB36152E1DA44E17376A380DE746F986F57B549B422AB71F4EDF0AA68F576D139A1A2BD61B1C9BA7427FFECA27ABB765237DB27AD332A8067EDB75E3D6F5321308933B835EBC6BDF3339CBCC4B14DC73EFA1626731715FAE3D57A39A172CF72F1398CBE2DA9FE3C2ED7B4907FBD849BA8A0890D466643475F94CA90E3D56A3E9BC473039D18D6B4C26C087F8E8D877AE371BABFA71E3FBC49E57D329F51152BDE51893BA166B020D8B5A37EEFB399BE4332CD59C6DA7ABC598593D9E36C226E5ABC6AA0AC86C9656DAFA26D66EBDE4BE7CA1B3A231A16A364659DB4828936CD24F5BBB7A5D71B9CBB15DB01085B692AD66BE766BB71F65F968A7C5DF95FCDC1BDAC0EEB71FDF794D689D9A2C2BEE3D0FB98C4364F6A4C6FB76ADB343A07D2FCB298C563132ED82A699AB052575CA8BDF8EB86FFEFED6C7BDBD6913DFE7E81FD0EC42EEEA65D3489FC94C406BA80222B8D7BACD8B59434078B0BC371941CDF93D881EDB4CD7EFA4BFA2992457166A8D1FA4D1B5BFA891CCE0CE74FCAD66E065859A15A4B4E02CD56F5AC556F268E96B3CE52CE48E261F224A78CE3D7C5FC78317A395EF5F678F9FC72AC06686F848EA4FD13887F89BFBF4CEE659F9BA715954FAA95B351ED217E38ACC6A3F161FDAE1A1FDED5C7E343E7FEACFE109F8C1F1E9A0DF1E1C778F4F2314559C74CEAADA7D1F4F175F4187F3EE884BD43FFEA20F5A96AF3E783BD36AB79F378B0D760D989F4A9EBFCF55966A8D4DBEB9A5595AC9FDFE245BA298B67F159154C698E2CC6976BD86A00C5FE69A3E96CFAF63CF9CF3AB57E1607B229437503512FF007E936A5FE78956EA5AA92D5097BF63FD038E6CFD1428C659129FD6AA4E680A7A7D9CF456BF7F1E1CF9CB1F5B3436BE705DEAE7C91C2A37D1E84B2B3B29A1387878742A8FD747FA0BA7F2056096FF53FF551F87B18F981FA5BE20E70DB34B25963991497E229FE113FC9127FF6124F55DD2467BE91789CFC88A72BDF102FB3C97419CFF571928D27219BD2973828D89211B6493650983DBC4ED7E5E9FB29DB0C99C48ED6EDDE1D83F4ED04E4DD837790DFFD7073C069B59E9DB95B221D050997DE1EE0A40E984C9F26D37855F12E32975007A8A8BD97A971A5297407743761BD3184F8B08EEE8F29775A0BD99DCD5A2219794F478BA3E7A3542FD2A1F9F23A7F8CEF67E3D7F703D26DD894FC7B7648F7742567655849936C8FB97F9B8E9E27E3944B6DD6326AB7D73268BDB50E5C8D636EF9B555A7EA75FC4FA15976919F57649FEEEF0F7B53D1992E94805CE7907F1E6710CBF9D6BF5AE2A07D203EC8E970FCC7FB94FB5173861ABED785EAD241EF20FBF968BE997B5B6A0D63F36EE63099A5E65BCBA8C4B68EE6CC61AB22C610298DFCE51DDF3975904B3BABB36B7B1B8CF2FC8AF17CCDE674358BF88DCAA86918175488A3810CC36840E5642CEA514CAA6986E316EC8A1785EA26AB8EDB2D0872A3DE5036A8A84522D820E6154BAFEB6A1E88B1BF7619BE2D9E668F2DF1A7F0BE793218E68BF14BFC4BD6468FF1328C976A4DE51F027A6D4F1295E4FCFB255ECA5977F23897B1E6C932B32FF3EB4225DA858AD28588BCBED3FEA4FE699FAFFE713CE21687BE8B7B96CD59B139570B6D6A65641EBFCCD4F2C8C3E4F1751E0F9F971F3E4A33A8379F97C357A59286D2BED3FBD1FC5E1605CBF96BFC49A980E1F2ED25966F20567790D752C0D5118BF958813F89F73F56AB82038FFF5ACBF8D7F2FD629BBF14426DD3602FD7920AEB6B6F3054EBEC52F8AC9D507DD0EDB96D2939276AB554FC19BFE1818B7829A6F14F311A2BA9BAAB56D68B09EB59707B8164EDB95A3CCDBA44B26CF23A3BBF75750B9E28468F81113130AECD8CEC0E9CDF93556CBFD7C6EFB140F92E1AC8C034DFE4A56F32F8EC082A392A8DDC2B8DDC4193B743C736705BA07FCB8E8C58919DA0DF1DF8E61FF03722D3F7F6F952E8466E2160B68DE1F5B9C73A342B2063986E91BCA3BD42F28EF6C06F13768041E0E62B3B0960B7135C7BFD6BCD0DEF36E0EBA264549EEFEFCEC6E5796A1ADAF1A13B6F6580CBA20B7FC72D0EC7EBE72B24AF9F5F0C3AFE55BB40B7B3C8C0EDA77F359383E87559DBA8BE2DD32F824C8D76BBAD78DCBE43C83EFBCE8D0ABD60F74C72FB32CD73B7554F01864A8B4519722A2DCC085CAFCC54B4E3B3D75ABBDEF393B763C34FDE7A0E481EF4D2021912C739F71DE52C1BD8DF39A5070E9D8A53751C8BFB1A8D3C5C7C6F0EDEABC06931916258CAC114C3520EA61825C9C1CD3550B306C2C1F54DE6081DBD41D9C9AC7250EF8C3C729032704839688304A668221223078D481B39486C232C076D808C618A94833648DED186E5200948918336609C1C34905179BE0439A8E773C841522C62E420350521E420118991834424420ED289901C24123172103FDA183948F71D42F6C1CAC1546858CAC114C3520EA619767230CDB093836906BF1CD4F3D96B2D5639A81F1B7E32B31CCCC826922ADC8A387E7158752AACE270C5C346FBEAE082E230C1B016870986B5384C304A1387AB6BA0E61094BBEB9ACC13483A83B29399C5A1CE19B9C4217EE0D0E2908E04276C1212270E0D483B71486A23461CD2818C618A16877424EF6863C421014813877430561CE6925179BE1471A8E3F38843422CE2C4212D05A1C4210989138724244A1C5289B038241171E2103BDA387148F51D42F6C18BC34468588BC304C35A1C2619B6E230C9B01587494619E250C767AFB598C5A16E6CF8C9ECE2704F3611C5E15AC495210E79770EAB949DC32AC3CE618251401C16DF394C304A1487AC3B87BA26730552593B87BAE1662797200ED9770EE948C484CDBE736840DA8A43E69D433A90314C09E2907DE79084C489C392760EE960BC3824EF1C26D24449E2B0AC9D43422C62C521FBCE2109891587DC3B875422461CB2EF1C62471B2B0E99770EDF811471587CE730C128200E8BEF1C2619F6E2B0DC9D431D9FBDD662178765ED1CEA3C87511C16D9395C8B387671E8F6BF0E7BE75F875FFC2B3E85988022E33E7186B54C4C326C656292612B13938CB26462E21A1C8EAF6D324B48690DCA4EE695895A676492898481C3CA440B243475D390289968425AC9445A1B1132D102C818A658996881E41D6D844CA4004932D1028C9489F964549E2F43266AF92C3291128B2899484C4118994843A264220D899189642228136944944C448F364A26927D87907DD03231191AB63231C9B095892986A54C4C312C65628A51824CD4F2D96B2D5E99A81D1B7E32B74CD46A279A564C6A3A76C118F961A4AEC0A716B74464EC6F0F2FB2A39864D84AC524C3562A26196549C5ED3504CF8EA2B6C92C61A535283B99572A6A9D91492A12060E2B152D90D0F44D43A2A4A20959CD565518B1486B25422C5A001903152B162D90BCE38D108B1420492C5A809162319F8CCAF46588452D9F452C9ACC682516894908231669489458A4213162914C04C5228D88128BE8D1468945B2EF10B20F5A2C2643C3562CEE31AA4519B68233C5B0149C2986A5E04C314A109C5A3E7BC5C62B38B563C34FE6169C59F945539B3B41682F35DBFE8D97255B284C3D08B053DF1F049D30ECF4AE8CA320122F73328B54338C2CF874FC0F9D25FD2067E472ECCB3D6043F936DBA02918CFC02912D7C0E95A451DBC647BAC076F656B86676277BE073D91FB2AF44C6C5674C003AE66C06169E02E0FD9F8546C2AD71CD68A664E3662F558038B20C7A28DC175DCF97EEEF7E5F9C197EFC3DE651B9D23E96C7CFE5DB311818CB62E4B58F7024B7F83C3DA1A6D7AD4BD3DB65A0382DA1A7C0A05359B25F6829AC68582BA1768E60FE569C373DF2A944D4028C82EAEDA2AC8FC3618C05B62BAFB2E5014A42EE1FA4069A0BF045459A7AFD1F56DAE0196EF7B17F9ADF48E78EDFFC635A02F7F208D95EF567E1094ED567E1FA8129196CAEF44E7EAB2EC4E04E08DFAC5873BB00B71DA353AE5C778D01F947E8DAB1E8FAD4C0917BAFF992153F5FAA55BAA3F28FF1AD053D539AE11DD969F70A3DBF23D37BAFD2FF8D58D6F9511B1E1F1A5DDEFD8CC1B98B2DF587FB114FBE11791FB2A58EC33A2031E700328F7ADC12750B96F4D06CA7D1A172AF7C32F39EE6655EBE7D2F0F34E3840846EA6E39962A669C80E21A608C85E82307986832F2C5730E5B870007D470F6528532710993A7B01D4E246BEDBF1A4B89B4CBB0C31424B717CE880070CAD685883EB608AB32543298EC40553DC0DB490260AAC54A2E9F80545280516C69B9393160F26BF249FB21E2A084BA2F89164C92237E714372465114674C00386B20827B8CB4306B2088D0B65911BDDBD5D377258772FEB148243E303F0C6054AA8426C97C826648E1B17A838327054DA400E1D43CEE8B9D5D0D299A19CC18A0E78C0E60D525E7097876CCC1954AE3967281A9FB832D08C0126CF1B86B298472C42E338298B81A49CA034D986210E43F7ABC7162CE9386445073CE0A6310E79C15D1EB2310EA95C731C2A9AC6D7BEF6CE2DC2301F668C1EF52DABEFC3AF9E3F1CF47AC0242759D94E3B48FC176005570F37CFD149FAD0AAF1D9D5DB3D6DF97E89BE5B6A0F063DAE1E38793D08A15BBA903D305DA0FC4EF8DF4AEE84FFAD782772661743C4F34C2E51B65DDB57D1C9850F1DF080CDC29017DCE52143930B890B4E2E91DED58651BF6B37C11880982C17E1E618854BF71D35C744B814AAA1632781486A33E0C60FFD055089C1345A1CC9C1EB65D7C4715E0B26074E74C002AE9815600130941CEC6D614E0E442E901C242DEB6E3649410B02C2351C4A1F9601E5859E79BF0B9CB393A473F4DA2E549225A99D025863733B805CB527F7CA6A3270534C017069B628D062835300CD45B750A2DCEB4E09465D81CDDFFD2E00BEB101E74D7CFA4CC433E155F7F3E4FBABE884C7870E78C0E084670B86273C5B3234E191B8E08457D5D4576ABFDF6ED6CBA321B348E71CB84782C682963ED3B4FC50CCB511473C76DCD20A507BB4E9CB1C05B09A9BB078C0C09739182D918E462217884649E3293FB520306E3AAE8A9BF3EBA107ECEB21E6C50DCB1B4417E63B2F3356832B8E6D3BBDF4AF32D8951C3B58DBC75745FA869AE843EA4D167833C8E1C297B6FBECBCA4A77745FB2F15EB81B68BFAB9303E1FD7D18AF879667594D3CF698D857D1D6E6C115F2F600AD0D74D6C92AFB3AD3E77DC934C00EE5E45E7773E74C003D64CC4A581BB3C6468862771C12C78A27135FF36B2CB8239306438B97D200B6850B9712F61906857BC94F9F0596545077EB5AD20DD2CDF617ABE91C341796C6862A00CE10DB4956B68677E6ECD73F842B5840EAA7605D44E225F286D895CF194C72B14545B68394155900E387E213A185685E8888A8B36966074199A4B8BAE5D20F0942FA794F98F56BEF0A1031EF0BE29C292B0DD72B8FBA50B890B26DD538DAF59DE1D69A0116411946F51B7369A3A56688ED2515568F29A6C4B64345B164933DDAE933CF9E7CC323810F9870F1DF08081ED0A7B30B45D614F867210890B06D519D702A906840D216079038F41AF5EE6C79FCE1E85B2960668BF4A97036332340D455A4223199C71A9A8691984885CC7870E78C040ADC586ED96C3DDCF73242E18864DD65A2B8F868F1E5937F0D45AB91D2B94B574D462B59689C868364AAD65EC244BFEF12E2C8303CE3FD6E86C4914F080A1CD684E7097870CE4201A170A2AEF626F0B310ADCDF4FDDEF27572751B7DEF916FC7E1DD4AFBF855FBF5B45181E6F0C37C909FDC14DC74BFE509F2D79CF8D255B8AA5AE779DBA93040B37AFE24A76BBE75D5B7081DFCB93E04BF7AADDF5071C6DCE5AE4A633B8EC59190460EB9ED3417515CC9A19C1AF3992EAA0B49F256045073C60CD6F909406EEF290CD4995C80592EA40F7DD769BF4A90519F3823A6578E9BB941F56DEBECCBF42B4229F77AEDA43F5D49072F0AAE1C0D75872D8D07372147D3383986F3ECCE123528E7ED479924B9BE29DB4E4628B36DE3C688F056E2D600577CBB1C47E6A2171C1D4D2E64A2D1A101443ED7060563AFA5E8291AFB82585BD420FCB6AF670100DDDABAB9E4D4A815BBEA35BA5448469D405FCABA81375AC9222EE022AA7FBE64D514C6ED5B93D476EBD0E2FF5BD53AF62B9D51E6DCCADF658400BB382BBE558229D5B895C20B74A5AEDAC76E2567DCF6954AB5EA552774EEA5EC53B3B6D376A4EC3B9B4BA810B813506D275D83BBF74C3A4D75279FB1AECBA7A3B8CFCD46FFEFA1013128DAA9911A18D18DE6D115EC688C330E0B3A1044696DDCD4B7218FF2BB07EAAF0175EC3AF542BCE69DDAF361BF59A5F71DA670DAFD6AC9E38AEBD7B0358B27B137928F7069864F7A6F2329DBE65A5ED3977310B669C1BDF3C937343DE57CCB94F2E6A67ED7ADDF37DAFD9ACD57CE7E4F4E2E2E2A4E29C54BDBABD734358AA73537918E7869854E726F38CCE5D9C9676EE8216DC776E42F30CCE0D7A9FBD7347EEF9B577E30F42F54803C08B77C7A1EB791FEC5EFAFAC5FAD17723A7360C2F816E6C0F4B16916EFFB743C7A9B86EE76ADD2154C37717448985EC45C472F63AFE23BE3F321F358F9FE2D1227D58A371523911776F4BE9793FE793E5329E263F8EE6A3E962DDDE1FF1FC8F78742FEAA74767E27F9207B565E92AC6B3E79779BC9078D91AD13C6AEC1DF33A1F2D27B3694BD4C5221E8B0F955AD3A9AF2F7C2CDFF878B4170ECBD1DD936CD3642AFA8777AF0F0FF15C2CDEA6E33FE6B3E9E43FE93EA40F1F980E4F1AB6E1D4956D57FF2468491BBE3E3F8FE66F2DFDA7CA1D14EB5D45683E8CB41F9E5436CFBCD57C18CD96A3A7A737754CEFEEFFE2F172A13BAA5AAF89B6DFF5235F59FE7934BDD71EB685D5CE1ACA44DB1CA33932D9BC7FAF82EE3EFE2196F3F12747BDFE5788E8351617F19D7A105AA5D9AA9EB51A355175AA15F8CCF664319E4DA7B22F93E9E3666FBADB15DBF766D345CB0AF20E901EB017AB7AC8D5ECA7E81C3C8BFB1D4CFAEB8A75D93EA7B4617B5AB209F0F932E52C278BE564BC10F7774379EE500DDE64293EC87F874B35589F6B679FC4EAAF5FF2BF1F6D98F3D9D3D3DD68FCA7F8A0FEB7C13A9FC4FAAF5F9F9D2435DB25A544EF649E489AD397894B66ACF774F0415D3D15B4F2B4F81FCBC973DC921E72E4548F946F88C39DAFFCF52F157968BFD27044207392F06FFBBD41B47DB7E2AC92A38208E9CC428164675BE26F0AE354ABB54AB37AD6A8FD6D7782CCC5BFA4E5C6B3FB383153B4D6C9727D4CB329FE8E7AFD3F504B03041400000008009B9B5752DFC61169BF280000D941020008000000554C4F4732315F31ED7DDB73944792EFFB8938FF038F6723D6E3CACCCABA6C040F75B5193060C09EDDA72F64218F993597053CB3337FFDC96AC920A1FE4A2DA1A6EB73B72F5CA4A2C9CC5F6556DE2AEBEFC7476FEEFCFE0F2A04F9CF0268CBEACE939AFEE3CEFB37779E9667E9512E8FC337E5CE9B9FEF7EFDDBBBB75FBF3B7AF3F5FBB747AFDE7DFDD38B575F3F7B3CE547DF857B0FA7FF54F8A7C7F5C19D8D167D959FFED7D367E5BBE971FDF8A9F2CDAFE5CB5FBF79FBFAE717BF9E7C9D95FAD39B9F7F95C5AB8F7AFEE2ED2714DCB9F3FFDEBC787E17083CFEFBFB37EFDE1FBD97DFFEDBFFFD3F7F5FCB1B2A065E1A6FDA39DA843792FFDB5756BC1D3D7FFEFEF54FBFFDFCF3C9DB3BD13CBDAFAC360AEF081117586D44F5383923C169E53E9270F1AF54ABEF6E48D68B976F5EBF7DFF29453FA0B1CE8B248E7F7D71F2EAFDF1D1BBE3A3E72777FF79F24EBEF6EEFDEB37AF5F9DBC7DFBFAED5D7F33E241F588073547BCFCECC92D6CBF1088365CBD5F56BCF98FBC85C73FA6A7293C782088C09D532040A98D19BD160F677FDBA3EFBE0B0FF3DDE3B72747EF4FFEF1FAED7FBF7B73747C72F6DDC7E149F8EEE9DDAFDE9EBC792DCCDEB9F0A1F2415FFFF5F8FDBBAFA717EF8FDE1CC3F4623AFEEDEDDB9357C7FFFABA9172F609CFBEBD7BE98F9DFCEFC9D7473F1DBDC1BF1FBFFBD3DF8EDECAF23F871FC3FC62F9FDCBA3E35F5EBC3AF9EA6FCFFF7BC5FEDF8EFE7EB41110FAF2C63B153C0202991E385AED149C8FE2FF9FDF4EDE09031F889B4EA99B62C56A74C8503018962F2A021BA3D6261943AE5EFAA077A2C7EF4EAE85E5EF7FE8EB4FB7CCAF62121600B4566CD183ED01CDF401E8F29F8F1F3D79D640BECFFEE9B73F7CF78D8F4F9E3C7E661EC66DA3FEBB2C9F94C78FD68B680EA3E747EF8F9AD285181E3F29B53C290F53B9DBECFAF8F83019F91EF33C3EA4D447450C393F7B147FA8C2E47A901AD75F08A8B184DA15208C6BC91A7553D10C10B2B72614AC3E566532FBA80DD86423C3362DD95F4FDE1FBF7EF9F2C5FB773FFDF3CDD1DBA39727EF4FDEBE1B14E58BA83AD268A08BBCFE80FCBDEF9A691303F1E3BDF2973BA71B00B7A521B7E42109033CC3C0D31FA2B883BBE5E1A2FA7DA0FF77DA5E9EB4E52F5EDD797EF2FEE8C5AFFFB1DE627DFCB4AE1CCC2772B889006A29398674FF4EA8CFCA13F91CE152028297AF5F4D2F5F8BBF9FCBB370EF41C95B95D445AED03058BD65E9C9CF031B40F9C5C42206F6C57B1D75D099554026A511236905E0B66C009B8BFFEEC5FBD76FFFD9A85F80F1939F1D3A851D079E90ED07D4D393129E957B0FEBA3C6FCFDAF94C220A0DE7B161E27D81CFA7777CEBCAADBD40EB676235B48E7FCA0BF3CB9F7AC3C78F4CDE77073724735D76AA71C7D744CD2A3870F4BDAA941BF09D94FA8ACC3E0D4ACE0F65CD05FDFBDBCDB6C98FCEAE8D55F7F3BFAEBC9DD7B4F1F7D551E9E7DE5F9C9AF27EF5F88BA9F25718E7F9110FF942A01FDF48BBFEFE66BC9E5A21C0C32DA39E7DD29AF3CD2C2D2376CD95DB90D2EB176FE54A1E5646FFE75F4D3F4F6E4EF2F4EFE31AD22FA6172376C0D5EDA746752B7089E7D1719DE2932B3E7BD9B56D4E9295728B66A5D9203938DD7895DB5AC6CCC9C12C66B9FF797805C44E2A683322B2474D44399D4E5BC0DDD57A8EA0FF19BF2E047931FD2377590BCCD258086CFDACC8343CA79AB66B3A7A76BF44CD2660D42FB93B4B928D3AEFC863661C493CBDCF234686BB6E8D1549F21846A35792E12CB6CCF8449C0F2FC452BAC4958334EB0D2511776CA39DBB765F600F70CDCEFFEF1E2FDF12F3F09C3C7BF2C006C2B029160BE0BF6C7B2DF37E2130B7DF1DE837BCFEE95A7BB0B7F36B74C7AEDB1FBE8CFFAC7EFE3B36FE3031FE237F1FEE1D8BDF5ADA525DED256CF257C4FD7D0ECB17B09A1C3B1DB95DF807658D3A42C18A72B24CE3647B231FA622416876801D1E9AD1EBB03174A3A6A43C659EBBBC7AF9E2B33ECB44EB2792A427F5A1E18A54C7249F33E90DFC9F3AF0D11D618B00F7F43573476B1951316E267F6B4458DE4EC56042A9F3E2F503180334D4B2398C846DDC4163373455635A14D953D15D0DE71CED923A9AD9AC8414B29735BA9098C3439E89847F13A96D68F6A0C5FDDB2D958F3B0B4FCB471FEEA32C525D60EF9E9DBD527E3D5E59ED633A993155B33D7907BBA460F6C4285BA29219B92C46B0AD979311212D35255567E0496A330ED497EBA83B226ABB5EB1D94F2EF213FBD1B70BC6737DF19B55A8370C84F5F25D30DE537A00943983C1B9D54CE588255A63A32994D25B6458BEBCC615F03E579B5919F0895EE9E5CE76ADE0305CA9BFB4438D710B9EB40F992E6DD30AEFBF061D792C27262E2CEF69523D97BFA8C987833D90D5CA969D44D55394C1855CA3E1934642946ED7CAE7292B321DCC7F0B7B36B2C83F566AE7D54D620102E2C46B472C06D620F11716991FDCD583B84BFB7AB4F96606D3AA949DD78C0D9FBA8A76B06F61D1B7513272ECC36DB6C9C0EBE92F1C5E644AA869095DF97F6AC0ECA604879DF69C29335031F948DBAC92886A24CC9D9B18D0E83A0AE1428CAA06D095B4C728C7B507620B71A94B59D5A8B386003E7B51A7593420E12E2444EE454E640BADD25D03969B4E21A6D31285C35648FA3D906602DCC2224EDA09BBE94354B1BE120D85E7D55F29435B317AC1DFCA12FA64FC691B7D4D7A7914F4AA16E228DC138D4469C6FCBE2E019974D71E081AB757CFD49030BF587E651B6D63B633A4554D47ACD98814339E00B80A389B4F5BED392296B0EEDEA57CAB4273F1E39A413EA26E76CAC3A4589DF4261AD3347E56DCE20D11D9674FD11030B6F579F571706569674D796F1C06D923B867BC476F50ED804A0BDEF64BC65CD47777F9876F56B58265E77EC1EDAD5B7BFB5D8B044D79D2ABCACB1B3C7EE1EB7AB5F63730F1D39B09D9285985275C6AB6CC93B104759D9E472CE15226CB71773E02A7C476D2CB256BD767559E347ACC26F9E8A98A57FD75578EBE0FA437DAED55DDDFE868E68F8DC908FA595E62F0AEF2253CDCB40DA894047CE490B75132B177D514CB99A1CC450222A88C126E594236BF7B20C31BF95B45848653A77CF498D1C8C36EAA6AC83ADC96B46B61C6CD6A05D723A06C75C9CDD6E6FDAA0903B56BC0EF226300760B013A3C89A8103D246DD844DCBC1B41FBCD28E2363B601C0E536F803CDBE549E3A30136AE755274D276BF4C2CA33A2D066139F48585B5A51ED66AC1D2A4F5F4C9F98C57DE83533CA9ADDD604AF329B66029321989AB3668EC52652C582352632696DDD16CDE65095A70ECA864124D1F387E0DCBCBA43E5E90B82238297235FF56695C99AF9890D7B5C79BA28D30DE5379E0913EA26671205907F1DA318AD124953A9A04C29D605BFDD36B3012B4F1D7521AF95EFB598C99A81E762ED18EE112B4F1DB059B7C9A29D3CA7ACB1E3559EAE6199DCA1F2B49BADE5007C770697ACF187CAD3676DEE81A78034EA262D2637156B2D87A23C14C83A902E3ECB7F72C6D0BE569E9C059A511B6F8CD3D8551B80112B4F1BA7220070D0CAD34558B65128697FC3B544B39CCAD3FC9E06644DE0CC4E043A727245A89B2A1BCCB904F2681D8AE35AAAFCA2C624161330FBBD2C43CC6F25D3DEDB997DACB0AD61B5B469421E78A35613E2C5CD80BA196B87FCF4EDEA931741AFD52791BA33ED775D6446AEE40A7553B1B52A55B12863DA4B759ED94366A0122B97B4C5719C43E5A73B2883D5647B83FC65CDC05D6C8DBA8974E4F6169D2A0683ADDE3BA8C4845178ABD66C379618F4A0EC402E5261EECD29264B4B2BE4B6668D8D4E13478BF3016EC4DAE1A0BC657D625E9FCE12A9EBF66FCFF174EC0636A142DD54AD8B45E5C4C9912F0EDA2BC5982C9662998ACEFB7250CEA2CC0E3C995E2F23F91D5FE1EDA32CD44D3A9ADAEE84C6E0D944969312BCCB595B725895DBCB910AB3908BC00C90878E07ACDDB984DB324E13508C1BCD1014DE967652DE90B7C35179BB1AD56058DB1EDCC44E1EA19797933503778437EA26CD5C0D7A2A8EB4C3124AF11125A814C2C0B6379DF7E3ACECC1DC66FDB9DE0B6422B8DD6675AE8019D5148D30882907898F4DA9E21051B1D1462217A3365B9C3235EE61D9C11C15901C979D491ADA9FBF0FB79013C56E36A99ED5E272CB37E4ED705ADEB64A59B73653D3C4AEBD03D7512956239F968DBA498913AD3824A8A614C8C107610C19338576E1628BF7A7063B2D6761066DB4EF3A45AC348E0CB3C689D869A33445A05C5C4B33FAE07375D1EA00C6BAFD3C2D6731D7882098F7541B61E01E9E46DDC495533446BC5EC62A7E5234499169372593139F696F2E4D89E7CF6BE71737293934AA77CB839737E457F8B57E23CF017169B3996FC8DBC12BFA622A25FF10A3EF04974C6AE0E0B25137D5582B259F55F029D84CA5BA1A8B375C532E1246ED8B57340BB30809BDC8A3673989D60E0F3ADC9CDA3E3AC4DC4646749590E6A707EDF1D5A94F84DA11A01EF9B5DB46DD14883D26174B716422305BAF5913C71009206D7736C6C04DDCF38A2352137F107B6D55B266C8E76E37F78CE619D8751BF765E5FB40FF0DDEA2597D5A570E9FBEFBBB9C9EEDDE1636E01CE267346D6F28BD91E35FA16E323EB4292112E51104B03ADB5AA2AF621581E47BBC97398FCEBEF1605BB2AB83398F5C156AD44D1E7DAAC20507674C8AC1922B35637668ABD7B0C5C9A5A3E53C5899B5388B94ACF801D4CB67322EAE6F022424D9E8F4E305E63C36E7ED5009DA894A91B2867B4FA6CA9A912B4142DD5493216D740AD9B681EF95D02AA552D4AA8003BB2F3D863D98B5B2CE50CF2B625C7B6DFD90F3F802E8380404EEFB2FF3F7D6F73AE77141A83D01D2D80EA09F103D3115D4C5B9946A261F8C5546BC7F061BDD161F5B193DE731AB38A4946E49812EEE43DE5CBF866744A35E5DBFAC7C9F17B5CBA75D4B0E4BCA79CC6F6114ED66A7B72EBDA19D381227AEE81A62A55035BBD8DE6B8F555C3871E29CAB3EEE69CE637EDF68871AB897EEB58B7BAF0FE0D3D7D0E778730BEC6DB8116F8758F8B655CADBF52A256237ECFB6944B7E3D68CBE1915EA26649512824F8994A3C06C908C4654EDEEAD8BFB53FF9F87192D7ADDEB279735035FAB6CD44D8AE590341139243480B1B8EC1D5A5654185CC1FD3C2DE73177C4F2CDD94041F0343BBEB9D5C17C451D4E31B9949DB2E231469B912544CCCE61ADA44D887E8BD30806AB10A08335372B4F316CE3AFC16E8AF3323C07D9BDEE6ACF61C5DBE23CBE9BF176F08ABE9C4A319146339B3A6E6B70D882FA29759377AC418E821A28638AE012E862A3FCC2A6E4698B0FAF8CE515CDC38CDE8AABE867C710B43587AEC85DA143DCAE8C9AEEB9766E48C4A1423027D49E00F5B053364EA99B32558F8663AE3E8105EDDABFA6DDF3E0ACA16E77CAC6C8158279C5D1868DC4C1B3C1DE6A8D1BB242B0B96734CBC0CE2B049794EFB372DCEDD3BA72F8F479C90555083A5BD8B7E80FDDB6A537EE536A2BEA7062630BA1891000102CDA5242216F5A7B9C7CD65E8EADEDEC1B26851E6CCFF4D91D4F40ED636EC94F39A2D296B509DE09DCB203D046D49472AA769FBA2249B935F3314E315420AEE1EC63439FE2BC8CBC00016C304362C5DBE2BA226FC6DB21E7F1E5540A9181E62B41AB35231F9742DDE4835736BAA20285EC4815535C30910C2656485B2C118C95F3E8C04C805ECF0FD85CADB1879CC78ED0B1245E6F37E761CF453E879CC79C50BB021C39732BD44D88D504AE1E826EE68BC0A3F3A89CA3D446E56DF7C9EC91731E1DC569CFA59BBEE2F0985D919B7B46B30CEC3CE77149F93E2B6A6F9F762D392C28E731BF8559C27B399A61EBD21BB62BF294BAA9AA6092E68C3AD5E882AD1442B2B9B0D14CA5EC67576467DF687606F46C57A4ACD9F530FB3EE642DD547D740CDE386BC58A5BC7DA97CA00013DC404FB33FD8A08D7E7B6444AA4ADC59E6EBBC53D3521FC6E32C063C5DBE2FA3C6EC6DB21E7F1E5548A994C3FE7E1C67D87E094BA4979175182FA8005AD6157B29C08E4BCCAED2D50DC9B3E8F0ECC46A322D3CB16BB73D9E243CEE38BA2A395F5DEBA5E46CA9DCB091F721E7342DD5080035A318D936A55BD62805B775AF59CA1C498DA83C6B666A6FDBD09DA511C640FE2307771D763E63C36F68C6619D87DCEE353E5FBBCA89DE61FA35E2B8725E53CE6B7B09690CFACAA11DB95DEB0AF2E9E52378184BC4A6510FF8D75A212BDC9044507472170DD62CD7FE49CC7FCBEB14E0374EF3C785E5EFD9FED06831D56BC2D6EE2D3CD783BC4C2B7AD52ECD65C173B153B5BA5E6078DADD68CEC450A7593458ECAB95CC91AA448DEC490E580617495F41ED5FFE7616EB3179CE9658B3D8F9CF210EAA6EA52CD908DCAD91B8979AB2234122858EF142957F6F3B49CC7DC38C5683BAA0D6A64CC1B7593E764BD7521D940A5F89210738A59551D59BEB2C58167A355082CAFC5B949C92987BE8FF3E25E0D23EB369894BCE2CDFD8179DB6DD3CA1FDC2B9A5729E73D758FCB9D437395E9F4125C6A1B9C03EF0C83D15AD912AA2B100B52CEB8C5DCDA605ED13CCCDEB5DEBB4E0E1A40E943856027E888E4112CCC4F2F59ADE14385E04AA1760538B20328F04E057C50A0BDE7AC3358453E94D82200A5C58EE93DEE8A9C571CD3FA83AE501C3B66856053CF08D4B037412F29DFE7E5B8EDFCB4DCB572585285607E0B3BED496BDAB6F460E0D456A36E8AC894AD92F8D7B757CD4CF512FE069D2A61B695B7FB3CD0B0398FD97D0348C62BEA39EEE7DFD45A48BCA8C16CD42B2EBC2D2E16BE066F8758787B2A2530B8B52A256257CC62907BD08C7CB746A873538A281EA433B13DACAB82A716E26BD6558BB980B2C5817263C5C2F3306BAFD09B5E9FB1FCE1916116EA2643CAD85CABA35850EB908BD2726E022545B6A62DF6930F7C5ACE622E12739678FE19D8B666E416B246DD54D99812A3A6A8638996634A267391BD1CDBBB8A6A6F2A049A34ADC759A36FC9944EAF0C5C782873219E0319BB91E770FE2DC53F346F07AFE8CBA914132AEFFA2A35726E4DA89B5271190C04489CA266D3824ACC3552423605F6C72B9A87D9309255BDDAAAE6C31D821DA1C3CA5AF9A1EBB3F2E10EC1D542DD5080035A31C629403146476B2141811A0340F2B91A0E31D9C05B6C1119BC42D0511CF4E04D6F56A4AC19F30EC1E69E110FFB82F625E5FBAC1C77FBB46BC961411582CE1666266F578FC06F577AC3BE8F714ADD94A367E26CC982D83BA8A82DE6942D9343E3B63C2A77D89CC7FCBE71866CF77D0CF00B8C85BDDBE86695F0B6B85901D7E0ED100B6F51A524DA5DAB5222768586B8E74D0CDD5CDEA89B7232BA28538A464F35A1B3D16A48CAE64054D5DEDC219887998D51847AF6016DF941ED78A24517E646DDE44235A928D65E1B9F403B95D152A6AA8343A5F7B442308379931818ED7B9395116171B7D224FCD9E802BDF0B6384FE066BC1D4ECB5B562981C1AF552911BB315A41A7E88638F22B438DBAA97ACC722CA6A0740A3E297291415755D8A468CADE4C9F998719BD716C7DEFB4A41DBFF3D58759A89B9C36A003DB9CA87290BD6BB389D1A61A543215F45E9E96B3988BC40C33BBCEC4215933F04C8246DD149D61537C2D01C545825C8D363A066F720043DB7C6F73B07A7A07678BD623757208B266719D781B7B0E74BE0FE88FCCDBC12BFA722AE55D0BC63AA33D90461E67DAA89B8C41F6D62A27EE9D9583C0D9986B854A108C2ADB7CC265215E910809D139DFE9CF467D2E2E39D4D3BF243ABA8DDEE9E7C665CDE1C6DDD542ED0990763BE8A66FC584BA2946E34C50A1508C6855AC21AA1A4C30E88175DDDF7AFABCE210E0AACDB68B3B0D594FDFD8339A6760E7F5F44BCAF781FE9B5484DBA75D4B0E0BAAA777B63039D9C568B62EBD810B418DBA090AD798A2ADADD5A4CD198A9C6DAC1C58623AC0ED9ABF61731E166066DF18EB2D51A7070FCDF26ACEEC69A32E23E16D71B7096FC6DB2116BE6D95F27ABD4A89D8C131ABCE581271CB468E8585BAC962A5C021041DB48F3A18E290C48D54165456716F7ACB67611621C9816B4C0F664703779F35EAA69C99A30BC9598A399790D9D59CBCECDF5098F316BBCF1602B308C9A266DD8B0DDCC8DD318DBA2952502651618C5981A916992A5351CE7A3157FB93D99A85998DE00CD82BEB3A1EB9F623D44D39245B5C7268A3A9AEE86C7C0AAE2A65C472FBB89FEF8E7730B7D691E935BEA15FDE4406A31137F20FFDF2DEDABE196F07DFF79655CA685AAF52227670CA77C34951C781CDA8703089A7EBAD1C958CD5276D82FCC48E4B16D757E7AD6650C73A2D3B30B7199B9A7B95865D77735F05334F2A2043ACB64626978A051B4B9673D2E9A4C1E07EBEDED0C19C751B47DDE98822C4E59D289637CA9F0B6F8BEB93BD196F87D3F2B655CA9AB52AD5C42E71B7FCB2070D0D1C7434EA26E30921E9A031B4170D33BB149D5236C71C74E22D061D839D96B33093954803A153572182819DA246DD04A9EAC241CE0041559B964600F0364874696DA9FBD94B3A8BB9480C1CBB9E234C5E2DEE7682D566A3C1C4C2DBE26A4637E3ED705ADEB24A596D71AD4A89D8BDD6D87B7254D60C7CF3A251378923ED9D4D6495AA94B34A10929C9725711BE613B6F808CE58A76507668F1282F55E89A3A1634BA14E4F8612416439F78B1C114E40D7C1A8844899E477FBD985308F799B02E44CEF35480DB4BC13C5DA8DE22FE16D7113EE6EC6DBE1B4BC6D95B2E7872B5C143B53ABEE75A119B86ED9A89B24DE70C69B44946275267231AC554E62434B15EEF6E6B49C87D9686355EF39698DBC5B7FB50FB35037E9906C625F428CDA562FFE50722526231B380563E27E9E9673988BC48C92E0721E730174EE6504D83DE62BEAA6105C8E81ABB063732DBA269B11643FA305ACE9FA5790A717EF8FDE1CC3F4623AFEEDEDDB9357C7FF5ACA4D45CF6E5D57EF4A4E8E1CC3FC842B719EF4E22AD4DEE326C30B85B7E5BD0D7833DE0E7ED1975229113B39653A97A49A851D7766E68ABA49CBB18FA92A319A54AD27D45E41F6A825E2CC719B8D3E63F945B3308B90C87B89513A30F3C0DD992BEAA69C49CB39A90C81A93A66528991E5F0AFED84D45B6CC21DD82F9AC55C242687A584061DCC3D2E2DD246C57A937EFFC6DBD2EE32DC90B7C36979BB2A85ED85A0B52A2562370EA9EB807A1CD98C0A75934A3A64092B0CCA8169A2373179EF8B4900C0D6F39E9C963D989D555A75DEDA93353BEEA8BC0266F253B5CA9BA2142207D4C96943D973F2550715DA61BF87A7650773AD143B34F30942510E5C5AFC85007E932A2E80393F18FF8FCCDBE1B4BC659502C27589B995D82D7AE80C2407B06ADC469F157593F281B42E1E4349852DDB5A385154356B1D15EDCB95B079984548603CD37C3F97AC817173EE8D3A9A944D200765D01E096D5529554EC6E954822F693FE7CE773007F25E39EEC496A897175BCA91B2C97D1AE14D2FADFBF986BC1D4ECB5B56A9A653EB54AA89DD2A06DF49D1C99A81CDA850475330D9E8A24350CA1461560515735289B40F39F92D962EC73A2D3B30B769C9EC3A4E11B21AB79F6B45DDE4B898E841595D1CBB6C591739120072A8390BEE7B795ACE622E12734DB77BAAED061E81B1A26E821C7D0E419CE0A4311304CEAC9D15DFC813E4BC45D51EAB3E8DC86BEF84ADA404ED01EE4E270229585ADF1EA2D56E13CF41785B9EC77723DE0E5ED11752A92676DDC69077020D59336EDFDE8ABA290753244C16364147F1885C2DD50744475C6BDE6643D7605ED13CCC869C72A69343208587976977840E2ADBAE2074CF353CBC4C7BB5503714E080560C71AACAA04DE2EE17F1677374514771FA25A2C981236F736AD2D893747B8A834E023EDB2924CA9A215FA6BD866784A3BE4C7B59F93ED07F8359B0AB4FBB961C963349B7B78559B76DCC5B97DEB817FB57D44D51D9126BAC502C97C24161343105601B4A60BF9F1582CEBEB1B695087A3E83595CAF36129B8DFA9985B7A5BD2A7343DE0EB1F02DAB94C060D7AA9488DDB7172D3A2D2A6478E0B6A446DD24E7496285B5181B42B659A5501378EBAC4BD5ED4DAF7607666E8555C39D0A81AC19B842D0A89B4C46A58AB7D9A8003AF98A5A40B6951D5915DC7E56083A983378B2BE17282C6FF62692878DFA9985B7E59D9637E2ED705ADEB64A795CEF808AD88DD5BE7B5A0EDDC4DBA89BC06354462772541C2893D03B30D1E48C801EF665EE7C0766DD8AABEE0A98F52173BC2374B4758AA8E7CB786D0E99E32B85BAA10047B46266E26CB576C199CA0051EC16BA0885A11A02B165756F33C71DC5312D0EE83CA7D1D6B82133C79B7B46B30CEC3C737C49F93E2BF7D93EAD2B07BFDCCC71670B7B65AC5FF5856F557A43C7C29EF50462E53CA612132AC2129255913270522ADA52F672B27667DF302BF0D0E905405E5E171529AB37C9AE7EC2DBF9FDAC17142F4E3FFF7AF2BFD3099E4CEF8E8709161B0666CD9E5BC9DC29A7ECFC71DBD6CCD8992F844BCFCEACA89B126717E5872C2E95D33529B654C17943C160B8C91D960B302E2252EC61DC2E37939A6F259335E79A303E448AFAFE83F028E9A25D7DF89D488E9F8D12295E8067F830B1030D629B7BD97923B0AD713361E21A7CF6274CFC44A85D01CEE421C7B05FE8A780C595ECC81716E365C53B62544E7E8F9A18F40DA6C46E64BFC68E117B5AE35B28DDE98D9535E7DE671F2846BC86374438688C7859F3368972D618AB8F9F762D392C2746EC6CE1D63ED88ADA5B961EF0CCED9A216C9F50371585C5BB100BB02F3AE5920BD8E8AAF8A51AD0DCE03D894D6DDFA001E2FCA6117191B330EB2D78F1E46171ED3720845F9D36BBC4DBF9CDCCCB0910A75F5FBE174500D99B6FA6E3E7304C8C0888EAF29DF733B1136B983D6E576BE6DA6FBE1034B376C64FA7D44DCED650B4A60035DA28DF60CCDAE7CA08A9447FFD54FC2524171126766066A5BD23987B06F5748DBD1C26F2FDEFBF8FDF5AFEDEFC085A7D47CF7E18244CBC84D0F091620F1D4BD6CC8F006E6BF0DC4C998B91E21A88F62752FC44A85D01CE5C4519C28AC9AF27472A2A7026C71824365428DE52B0D64747355675FD4CD7C656ECDD3F5EBC3FFEE52761F8F8974191BD88A4E88CA739DFF8740DAF3366463D7EFCFD63C74FBF7D700FFFFCF4C1C1986D011D6E43E766EF459CAE994B7BAD81E860CCFA021CD198B9A96231189DF20EABFCC82A73CC88C055D750E1FAADB21B1BB3B1335F3DC571E8B49D7D2069B506D48899AFCDC33C84B9D4DDAE335F97956F93DCCD5AE76B8D11FBF8575C4B38CB498775F6759BAC26BFDF341D76BB22A5A1BD3EC2098D84AA41854C0E0A589DB5B1995D31A6E5CB84BA6D1ACA41D3649DCD44C8AADDC1EB604E3893171D0273A16EA29C53A458E5FFA03878B6129F976A6DD02A72BA7E5EF40F7338F29A69536798B67B97B3CF68CB37B5A6E5BD4CCF1BD4391B6F4C4B9BFB7F43DE0E77496E59A58CD1979FD238133B33B1E9A914EFF89A4FD79436EAA668228051BEAA0C142A41F5A92AAB3438CF5A6F7120FC58A9DF0ECCDE3863671F435AADD174395B72B84BF225D0D10D1ED9A75D7478265BB2D777493E116A4F80A3DE1F3FB562ACA6E05CB4A6A2AEE2FB935229256F8A844D3AA9A0C1EEEB5D928EE2B0F2CA3A35D7507DBA66AECD66B7D992CD3DA3731D1E83654B2E29DF26A1FDDC6D88D5A75D4B0E0B4A8C74B630696BD4AA7777BBD21BF42ADD99F93393A1E4AA0271E1541B48E46C62A801284686C4E5067D42CBBF4BD2DB37A63D0E8C73B783DB1A3797F71A0273A16E12873DB20BAB2B93143CD800CA7B9F4DB0C605BEFEDBC00B9DBC0C12AB5C6E093BC3508B0BC073D3A64ED7D0D2F2025699CDDC36474BBB0F751DDE0E939777A252AC8C13A5EA4233F271291C4C25702E2C3CEA50B2ABA159504B596996753AED4DCEA303B3516C3C774F487D98BCBC2374B4426A358F2E3A73ED6EFB9DF3B828D4AE00877600354C35898F1F55D618ABF7B5E604E88DAE081A12DDA0F0F947C9797414079501987DCEF374CD909397AFE119CD32B0F39CC725E5FBACA8BD7DDAB5E4B0A09C47670B6B8B8AFDE74E5EBE5A7A834E5E3E337F76B24E8BB5CBDE55131183D61E42AECA39C811586D7164E8C0398FCEBEB106B4B19D3C175F78326521F122E9ABDF2F5CF1464B7B854878737C7DDECEEBAA39C4C2B7A052E4DC3A956A62376D304D27BD246B6662E12F044DD78C36EAA65295CD5AD7401E8082C5DA72898E4A4DEC54DA62E56CB058781E66C3E88CEFC2BC2E16368758F80BA0A39571DE42C7A5973573F5FF3510ED532C7C41A81B0A70402BA679CA003E251B8C029FC5CB8990426048D61A1FC525DC4F67705E69D8A045980D0040B70EB0A5394CAC7983DB3F9FF2767E3FDBE5384CD38BF7476F8E617A311DFFF6F6EDC9ABE37F8EE3320910FAF2C63B15BC432D4E7B179C1963F385C0E9189B53EA262381A65259D512948A417BC7AEC831648CD39872B97EF7F9652C97E1347580664F60CD6C9DA8AD813563D5ECFD87DFDBEFFD9FEFFF588AA3F804BF1FC4695A83D1F86ED33C3E8000ED22480F9F734F6B5E749BD680B4476ED345A1760508235B3284A9AAE2126BE74C561AA3222F5E1315879CA00D87DCA6251BBC8830AF3A08A8E40C9B8D07576BE6FA0E775B44D8DC439A7D7F72E745844BEAB7491A7C8DC5FAF869D792C3828A089D2DAC0D1A6B69CBD2733473CB7E080328D44DB6DA1C331B4BE85376193C8957A799902476CC37B8657F2D03386AE438BB734466CC705A7E9A411DE54F7F403D3D29E159B9F7B03E6ACCDFFF4A290802EABD67E17182CDA17F77E7CCABBAD5129BDAE0CAFD8A9F8FA5B0BF3CB9F7AC3C78F4CDE770737247C93FBBE5E8A35EA6470F1F96B4DBAAF00DC87E42651D065B77417F7DF7F26EB361F2ABA3577FFDEDE8AF2777EF3D7DF4557978F695556FEC0B51F7BBFF5CE9F3F12F12E29F5225A09F7EF1F7DD7C2DB95C94439B4A4EB377B8650D032D4209496FE490083F66214AB83947762425BC09D9FBA08417E472510EC63A9C7D62CC83518CC34E193AA56EF2AC3CA81C3D32451351991824068CBA98E0F34D02C03FC6200534DEADC37D253523EEA199BD70DAD6D0B0033F4FA99BC832A053913529952BA24EF2FBEC5D2C90BDDFEE008DD33979E3B8BB1DB04919F2B3B7C31001B43F773B6CDC93161DD39559AB337EDC224EDA8D39E2F6D70F73D2DE8CEC3FFE49FB895C2ECA417BEDE79E419635A8091711738ABBB001F48D9D8578BB9B333492B3BB39D57B15705E10CB4531606B209DC978B7A194CAC2FA62D1B749DF57C65B6BCF9588D4F6C47555620BE0CA2EFA337E2E257E6759F9012512F04370749103D59C939900A5AD01737596E0C177CFCAD3673BB69B466DA4B18DA12B0CE775D9D99AE5BC064B2399CE1B91FDA9EDFC1D843F92F1BC28988B821055649E0918DB1AF457BB2F43E8E12659A2338696A287D76069A8A4F94DC8DE073D5C97B0FB5D10ECDBB352B3C2328ACEF59E8D1B46A012ABB201F88D9F6594AEAEC3911F470B6F46F61F3F92F8442E17E5C0464EBBB9849A7C5F4279BABC69E5AFBDEFDBE633BBDFAC1AF8EA2082247052B046FD3E70E276CE89B7E6CADB732B46CEA7572E31829BEFD8AD9DE2B22FAE7C3972C5C9F981AC9739D9FDE6123AFDD50EF6959CC00898E0D56F1A08271A9017F72C9F331BA1F4096FE74B416E39570DFE35BD7A3DC99F7FFE72A06B99ED46DEA523E64CE20CDACCB528ADD6909AE9CCFD42A8CC15E890A615756A026F2B87A858F9184256D57B150B2A5590B34ED72FCC9E077119970B661116F41439A3E6325BA76B3E06A0DF884D1102E3BD07F79EDD2B4F7718B55D64A94BFE4CE7C0201BD44D3E5A90285F991261F5266DAD0C086D8BD640B49D0D3A76D3406FC73AE5AD9EBB437CBA06669A06C6801CD4A49C206B0B5BA36B2D013DE7DC060FF94298186F30717233C847ED929D875B7EC10AA97304359FFAFCEDA7F4EDFD0FCE1B6D1E19DCAE79F2FEEA891367C4F3BCEF790DF2B7E67B5E83930B8FB6FE253E3ACFC7F515EF62ABBF7CE2BF87A7FFF530C92FB6CAE1458E64F7F1DC53B5A76BCECD2B18E674949383363A1D85FC81DDB746DD842961A8A17A93C55729CA5A714B8D09DA51CE996E309B6D3353F9FCC5CF3F9FB46B07E398C94F60BD08235A0F38F36CEFD99A99317CFB0DF56F6F9E4B38F7F3DBD72FDF9EBC7C3DCE30E61ED868C9EAB9379ACFD67C3C56CE98BDF7ECBC3DBE08FB0FB0254BB519F47A523619ED20B361A7D0C676B5CA8843549C6D13186FD721FAFD535EBD7EF3DBBB8FAF0C5F35EBE5DC279EDE57BEB891C4B37A2FEC8BBD6806E3E1A38753BBCD3C9D0AEBAE2C94AFB6B53F3CBC97C2B37BF2FDB36F9D9ED9BF1F77D3698FC4DDDF2FB8B57E89056C496652A8E72A42A76B060EC41A7513B94C5CB8462BDE39BAF6630A9E322786C0C56C2D105BDD6F3B8BC79600B6D3062DF7EDCFB9E2D283121E7EB8F1BE238FFC1A2E11CF8C3419639FB29E8830A281A0838B21269D952ECCDEE65029C77A83A1DB1BECD3662A25705CCA19C99AA93DF53A0BB4C7F3634DC72D518384C6B0C1C66DFC2CA3447D1D8E466A14B911D97FFC12F52772B92807070EC0FFDBFF07504B03041400000008009B9B5752E5A69501B8010000460D000008000000414C4F4732313038CDD5CD6EE2301007F03B12EFE03E40AA99F18C3F7A4B69B48BA05D4AA2D5F654B53479FF47E818B4BB04820A22A1F866C9B17FFE3B1E97F9629601609E4F9FA655BE98A031E60F9049AD36A0CD1010021121729468CA7CF15ABE9455F1A8435A1DC2F0669BBAC9A87E5B65FC4E75F6CEAB55061F819BDAAD9A268AE96AE35179C0718B0A28761D0220E73A5AB301478ED11D7674E6416CC95D411EEAB0D45B1E69578138B0F7E3D1CF09CFC045EFFD164F1D770930DD7108787466F6AB583E15CB6A3DF455210F60290677AA23CDC67A30E2F8B0E3369FCFCDEF0E47F806C78F7D07C137386E2EE7D8FE4FE78F55515678F8BE083ABCC07DD97674DE97E438BB8EB566F31A894838350F8AE4AF210F8A16FACB43771544EBA9A513EBA9030BE10AEAA93AD0F69647DA955E96481CC7235D7216D36AAE7D6F3BF2D0AF2C319BC93D6C1AB36F77CEC8E39F237CEDD01701D00FEBD0FA748CC3920CEC38E25CD421360EEBC0E3F218DC6177FE8FF5FBD2EC38381DCCFFA5C94BBB3380A3338FE490CB3A3AEAC7C611CF75B466F34022A8EF4B8A7FAFFD3D97E77D07437B685F8E13EABA457D0FF802EFDC17757DED90BEEAFA66572160408C9F504B03041400000008009B9B57520175BE318D010000A60A00000C000000534C4F47323130382E583032CDD54D4F83401006E0BB89FF618E1A43B22C4B0BC6CB968F4A2A1F02357A6A285D08B10542F1E0BF174A1A4BAD551221CE690F3079F2CEB0783E757D4836396C93345E33F8AC6784F7478C308F30C6BC88C63C81562D3CEAA848C0B234AA1EE4A5408858C46116841C5962C62D491872682591888DC22892C5A36E88C8E2885C5E783B095555F06D98CC755D730F2577434A6C07FE85C4B074FB16823C867516BEB2155C05699984491E9459F10E7991C545B08198A5AC08CA244BAFF7399AD4B0C0301DBB3AB7266A0C9D6327C9B8EF1CDFD2DF2759BF05E04FA83F05C5D5A86FD816D83A68D6CE9F36C6342B21656C55F5FC527FFB6D98D49DC193E67A95C33B4CF269B8241B49992F625656A195DBF64C0795D4DBD54922F7BBE71DA6D39BA49E8EE3DA53979ACABDA6CC6E1E6CAACE1D7527990E2DF9763A834ACEEEC90909467D4ABA4CA75FC999BF3EAA6A2049BD27F54DEB9A50F5E4BC17CFD7CCFA9AAD258FAD4C049E60221CED8932414DE1B1D849D2741B232C8ABCF4E3C60E2A39BBB1272404F529E9329D7E241F504B03041400000008009B9B5752EC9F64BEEC000000620100000900000044445052482E786D6C5D90516F82301485FF8AE93B5CE84CA6E452D38C266B2648DACECC27D339B699289081E2CF5F4125CBDEEEFDCE393739171797E361722E7E9A7D55C624F4033229CA5DF5B12FBF62726A3FBD195930B4CD25B2EFB6FEE774D9B2899C1893EFB6AD2380AEEBFCC6D6FEAE3A421F700E728D9FEDE154340CB9527CC3304972F5CC30572B93F15430CDF36D229672BD056978FE1482232F5E10849CCBEC8A10463B723D7DD5420D39BDD146A40877D68B093782D180865E403DFA308803EB4523DD89701ED15934A58334105CE62B991977E1718E302E98F2372DD66E1208E38C70AB00B746F0B725DC3FC67E01504B03041400000008009B9B57523F481178EF060000014200000900000044445052532E786D6CDD5C6B6FDA4814FD2B237F58116D12FC1803A194CA01BA45490802B2A2FB25326648DC1A9BF59824FCFBBD633038603386D68E5955AA54178B39F7DCC799E331B52F6F530BBD10979A8EFD59902E450111DB70C6A6FDF459987B938B8AF0A55ED3E95B551FE9B3AD4FC2BD36ADC27F7E169E3D6F562D165F5F5F2FA93EBB349C6991DD009F1096B7BFE8D69CD07A4DEBF5B4EFF55AB3D9EDF5EBB56EEF7ED0D1EE5AF5BED67D6CB66EDB7F3F16DB03ADDB908A70E5E64214254D6B7796976AC5F5C7D9FDF7FDF6A07DDFA9C3F5D0BF6AB7ED4EABAEC9A835E83D6A0FFD5B24627128C8A22C89B2AC084890AEE40A96852A7AA0C44502FBE6FEF7FEA075279CA331B14CC0B84042FC32845AD1FF0EFF6B1988DF88458EC2A284B0A8326A2659232AFCA14F679F14E12F621357F7C8188D1648BBD6BA683437ADB170860C9704D7E72C12FE0D72381E29025538401555425AB78DA6CE98A082A15B16AC74E23A534417D423D32A1286A22C7CF2975E7DC7E2598ACBC65C7E940D3F500433C726B6B761AAFDD878E8F55A9DC67701E9949A4F36C0F29C0FCF3B958B0B0F598C6FFEDC4E336F312348683CEBF613415D931804DD9AD40BE5D7C47137F8F6C7244584A5D34CB872526276EA7F45CC3ADE39E6A6729ADC5CF1B8C1D270DD0C5CF22FCC3F0F455551157DB5F4A72798B9C8992063594BAB1B28A29EEE32AE0A9B2176BE9E62E799A195C4D364498AD407619ACA65D45BB14389450C0F140E9B8A1408B1D64440FCFFB97E68DF3603167C062460E05D11EDE33955985CE950C2CB118F05680FDB49366109B8DD10022CFE6DD29EFEBF920E115360A542B2ED2752A4BAC0E1CA0CC9A8E32B93354C8BC4D666AA74F3948824276E3EA66D587326EBF3DC7D780A25AFDD87A73BB02A0EB5B9E74C75CF3420FAC4F849D178EEB2944B424C9A6BE7890F459593E69833FA018D1559D0149620D78965DA21DE3E38C94E5488483C2552BA12D1FD3603B03E4515026640F6B926A148B7C74880BDFD4FB2585FF33F9E6AAAC991E202871160C4E616715DC7A5E7FE125F75D78632A1903628D5F8CAD112A2155A9D84EEA10A2021A00EE8DCF258C57ACF6437EFABC1B00CC014E8D9395A8DC90D28B81A3F75D3841A29234244A8C947E76ECD4F4CDBA4CF1F3F4D649EFD8055E9B8AE4CEC71BA85C29BFB187A7268E58EEB2E056D5E868A1C39C843765DA9220F1BEB65B34526D2D648378C4B66A5EC66DD4727DB895A0F72E4F80FB70215A39BD59858207D3C5E7A596C72B0E8C34A7BCAA08706DAF503EA6A035179EC7F13A2C9833F6922E1CDF552050F1B7EAE2CBB374DBEA703DC3FE634070A598E1401E1115A51D17B90C1D281A9C6FDD776153D31B378D3F00AADBBC6B756E3E62C6D82149EB78055E517BA5ADA4D59E15A0CA5F2414E1064D336061F20C088DA4A0339427F6E1884523485BF74B89125B0E53C5DA64A1BCF7390E4C41B84BD04EED30D4A8635A6F074832C5D2505EC128BE894ECDB60670A8D272C723AA5149E3120CB6252C11A50B23FDD5245C3D30AF19EF245642739DC53CE34E9F80E433E938E2728927BCA1189B8CF51DE62395590DCE71BBFC5518E7E701538CA918FAF3EC454C63C8B22DE543EA834F79ACAE9761FCCD331F1A6F20EC4634DE52CDB0FE6E9979CB61FCCB72F726B2A63BECE88D38C3B3976B4A99C69929DE8930BCC9322B1A67229DA5496B2369531CFC0F04D65F9834C651CAD217ECD54960F3095B33B5D8479D644BCA99CA0E6F3B23954F926464E4D659537F7F36E2AAB9183FC0853795B8C1E642A679A6C277A80528D1CFFEF4C6525642ABB64EABC040BCF9FB1AC46CEF6DF63916FA762CA48B8D3FE008B7CA78AF658E499160D5710E4D72257799643BE2D72956F26C459E4D1BBE713B1C84B3C51106F911FB6A5CE8B0A2AF1B4842C9593693D5FEDC61F625E9F61A64B5A2773CB5A647DA89C6F20FCBF36B3A513951D259EEB10BB99AD446F66E5AC37B325AED480CD6CC759ED658BA16D6CEBCDF3BB3B6A36DB8DE5784B37D69152E2DD6E4F895B29849ED850E150FACBA82FC3EDE9238BAC830D026A44A0A1AFC40734FC54E1708F41C4C119F8AB6E2D575DDD34BD1EB1C88B0E308BC12500734D9EF517D371D3C5122922922451A1E3D8175DE24E4D0F5AD219BA73C6E6C434F4A5C2601CADB0E823D332BD45BA30F87EC25E4A824482C6CAEC1403149E33252ED261A2A0C2CC753CC83CFFCD382066D06BF51B67E9BEE4C37BB4104B4B9350C33567C1A696A1A1CEC4834F90D0EB3C6CF09869D7499937F88FF1B3C45CFA59E5E30F49FEA29F85339CF5E5C8599FB5A99229E2133D0351E69E81883D96B2938F098EA5644AC9891E9E2C735FDCC4E5A0A16DDE4D0F9EA6B328AB40469F91C0F6EEE46DE6B85E1ACFD2D38C01F794044C847D315031C400B42B6B2DA719019EEB117F3829B630F78C0788D70E9AE2EA07178AE11F6128063FE850FF0F504B03041400000008009B9B5752A85568E016010000D10100000E000000545244454C49564552592E786D6C6591CB6E833010457F05B127035D556870E48253590D0EB24D54568824B48D44200AE4D1BFAF218FA2D69AC5DC3BE74AA3314E2FBBCA3A958776DBD481ED4D5CDB2AEB75B3D9D69F817DEC3E9C677B4AB0682F7EB12AF67F4893AD5BDF0C03FBABEBF63EC0F97C9EB4C57EB26E76D0070C615FE3A7A23A962D412A25CD086A19B1395F3269FA68BECC058D1901AE69127AA068F2E6B8AE472917570BE10161C4542879A205792DEBF25074E5C65A7D5BF48526D6EAB8AD36067E20A8B3849134E142210C3D2A4D75AA8060B810B33CA29A11D73C6728845FF70A681EF7803FD46D3C78A86498AB4C69169377F70961A4875938E74C68E2F5B19146B1D07C96A78A49B88B7F5B8C6CB3F7E858703B208C8F0AF70F223F504B03041400000008009B9B5752EAC1A9D03C010000A106000011000000545244454C5643484B534554442E786D6CCDD55B6BC2301400E0BF227DD7B4BB3027311293E3CCAA8D9C84329F42E7DC266895D5DBCFB7E236C487BA873D244FE1DCF80884433BFBC5BCB69D7E15B365DE0EA24618D4A6F964F936CB3FDAC166FD5E6F061D46B362DFCA5EB3D54565D99B17AD32D90E3ED7EB558B90DD6ED728B25563B25C90634359119CDAB7D97C332D18E5887CCCA845098354F4630356322A07A94BF8101851968F44440C1FC5F5308C3857C92944C96F11157D10F1E92E744F517216A046A375BA9C8F2C3C9EB2F32C44873C91DC6A1C1346C905E3DF54165562DD3165B9895D39D149A94435F4C60F28BC2863211150ADBDF543FB0409A0122E8671B5F7CE13AFD6D2E9EE7335F6DE0FACE5DD0138482C5E79DC075FBC52A1B312D22B5FADE98917C10807881AABBD8F7E785391021A87D0ABE446E1DFB9E47B2190F325417E160E3B00504B03041400000008009B9B5752250DBBB5BA0000002B0100000E000000545244454C5652454C562E786D6C5D8FDB0A824010865F25F63672F42E649C58AA0BC94AD64DEA2A36B303E481D6D3E3A7A4245D0CCCF07F1FFC838B26794DAAF8AD9F59EA30CB30D9244EA3ECFA4CEF0E2B8BDB6CCE16844A37B6BAA8FC8F6CDD54DB6DE8B04751E436405DD78656B911650974424BB0AF5EA957196B422E043F114AB15A7BA1688770E5856729367B2128E0FE66669A16E7EECE95DC5F5A08A31847D8F40F1B900EF702294F3E6DBF727F6120B93C047444E8B74EFAD580BE1A8CEBC2F03A7D00504B03041400000008009B9B575225D773F5CA0000002D0100000E000000545244454C5652454C562E786D6C5D8F4D0B82401086FF4AEC35723FEA50326E6CD9419292758BEA125BD907944A6BD9CF6F25A5E830CC0CF33CF00E0C5FB76BEB99DCCD254B3D441D825A49BACF0E97F4E4A14771ECF4D19083362F57EF74FE475A3735AE3D7AE85C14B98B7159968ED1B9B3CF6EB8122C813EFA535F1F89E120A4146B0E4AFA9370296D71F0C3E556C9E95C4A1E8B68DA26840A11CC0225A23105FC7386BA6F468B20F419619430D6A503D6EF59B0812A218C955A473CF8E8F506B1126A11F315E07AAAA46F105C87C3BF8171F33C7F03504B03041400000008009B9B575239AE7843F70000008A0200000E000000545244454C565441534B2E786D6CBD92416B83401085FF4AD86BD1D15B9175C3A07B101323EB26E0496C6ADB40A2D235313FBFEB6A8234B9B6CB5CF6BD6F061E3374793D1D1797EA5B1D9ADA27AEED904555EF9BF743FDE99373F761BD9225A3A5BA7AE55BD9FE22756FAD3C6DFAE4ABEB5A0FA0EF7B5B95ADBD6F4E303468828CED97F278AE14A32804E68C4A11F2D54E6216331AAE7645826BCE209298062E6498C696E3B88851324A14EE90C1B75ABFE351116C85E049908F98310D2645BC11820DF35E9ECC9BEC114599A70C79124E9EF953C1332E8B1025678E7E96290A33794264B41E10CFD40D30228579D87F0FFE6CDE63F0546CFE38384C9B87F935C0EDB2D80F504B03041400000008009B9B57520CB80CB33D0100002A0200000E000000545244454C56554E49542E786D6C5D51CB6EC23010FC15946B050EB70A2DAE5CC7A8118E636D0C2A272BA5B445828008AFFE7D370F02AA2FDED999598FB5F072DD6E7AE7D5A15CEF8A71301C84416F552C779FEBE27B1C9C8E5FFDE7E085435E5E47F947BEFFA7246F518E881C073FC7E37EC4D8E5721994F97EB0DC6D59652045D0D8CFF9E6B42A390844B1E0E030527A3E33B1E310E9B93722519CC54E58396499B0D37E180E85884DD302D6896A796DBCC9632F6788CAC845236BA7AA4C626C9DE16BBF3C1D0EF4ADDFDE6B4A928E80CAE751691ED26BC03A085638F9A6D55C69C62173C2CD322A9CC8A69E2834BC8EF8549D7BC4071A243A9F507C762F51B7608269D2962EA5223612FDED318B2A52526529569234B1DE2DACE292D27500245D93586B15F17722EE08326BEB49D58F42601D0469A706EB2EB51B40891FB6C0DACDB0C76DB1DBE6F91F504B03041400000008009B9B5752B152EFE1DC010000830E00000E0000005452444C5643484B5F482E786D6CED975D6FDA301885FF4A95FB6027E3A28B8C2BCF7E191121A96C97C19595756CABD4025A68E9CF9FE38434025A459AB4A98CC817F639C71FF2F32A92C9D5F3C3FDC5D3E25771B75A0EBCA087BD8BC5F276F5ED6EF963E03D6EBEFB97DE152579F11CE55FF3F55ED2CE5D16913507DECFCD661D21B4DD6E7B45BEEEDDAE1E5039C126BC6AFA537EFFB828286152B239255A8A64CA476333A2C4F68C96E34C4AAAD8F5D8C738602C4E63CDAE794050CB267C047C6C523601CAB3614C504BA84D9E304579626062B2849B76CC594469A66F14CD08AA7B2495D910A4CCA4A298A0D6C8395F649A7E6E9C7A546F76A3C01DDAA8B9D230D96DE4E47223A98D601A6888C3C0C7A11F7E70BBEEE43AA2637BFCE063145E46FD26E04402A9385CA1119DBD37BB9108B75776E4848D5A050E567F51ABC0DEFA2F9AB3816B107456E9D5A0BE9B49260059D230D3B1B01DD466FEE7FCB58C537B8DA5C994656BD71222E66F9544192E6BA29DFB5F6B02B5F163FBF9AEBD821FE3C8B5E3F8D1BF430EB3D85E72CAA10BF783F019FEBB86AFD9A7040CA45ACEBBE03F123FFD02E89F760188D89E47C0B4D37FFF48FC3D1440FF5C00AF16800465FFEB259A4E0550C6CFE84F02FD944F412A2361D885FC41F84CFF6FD147F5930FB59F8168F7A4A4BF01504B03041400000008009B9B57529E26DAC047010000980200000E0000005452444C5643484B5F482E786D6C6D92CB6EC23010457F05655B8193B40B1A1923CB3125CAC3C8362DACAC94D21609026A787D7E8DE38408D8CD9C7BE7217BE0F0BC59778ECBBF72B52D068ED7739DCEB2586CBF56C5CFC039ECBFBB7D6788605E9E83FC33DFDD38756D51065A1C38BFFBFD2E00E0743AF5CA7CD75B6C37E052A01D4E557ECCD787658920E61CCF11943C4CDEC93856630475A4248F19E748E049FCE4BA1EC65116493C211E042D19923125B1CA704A1161A3088216B02249B040245134552C21AA6D33121412CBA9400C021BC18CB311E59C71815C085A99513E7896BD358ACDECB0A9A0666925E642D2B41E64F06510972AC49222DFF5BDAEEB77FD6733B5C6D62223BDBEF71AF8FDE0C5AF0D06429A85F71D1A68E49BEA0641A29FECC1860DAD0C77DDAFB432DCF4BF3223532269886615AF12FB36290B29D03F4D67320A7500DA7F0EEC1D80F66D80FACED03F504B03041400000008009B9B5752EA7417389C010000460500000E0000005452444C5643484B5F4B2E786D6CD5545D4FC23014FD2B64AF46BA810F865C4A4A57606E6B97AE43F165998A1F89027108FC7CDBAD93C5F8A6D1D8979E7BCEB94D6F7352181D5E9E3BBBE56BF9B45E0D1DAFEB3A9DE5EA767DF7B47A183A6FDBFBD3736784A1280F83E2A6D87C72EADE5539D0E2D079DC6E370384F6FB7DB72C36DDDBF50B320DDAE1D4EDBBE2F96D5962205292050625FD684E67611E62D02857321452E29424E1A9EB7A84043C5024A11EA0960C74C668987312334CC52400D422AC28338E69239802A4B8CC79168F99C4AE59FACC1605713A4D445A4B2E205B1A3AF0B19239C9D228A791C8FC5C44B43268C1E85C60D7F32A4643C3CC89F4BE1CA2D1ACA9871AD4FF406735528B045F570D0601A12A101C134016414C52C5A4161956649C69EB91B062F51E09516E3F4F678DA17EA4641AEBDBCBBE92806A0C627CC1A8B287D9C2909F4E69189807ECD22054B92619A71AE9EE902DBE9CDD4A900AA92611996AB7DD22C2CD564FE60BCE8E73561508194CED4B307D8363058A5DA9C0D7CDA81DA5BF8C55EF4763D5FB9558F9FF285627DF8895FFBD5821FB6BA1F64F869A5F11BF03504B03041400000008009B9B575252F5D8F07F010000EF0200000E0000005452444C5643484B5F4B2E786D6C6D52C96E833010FD95886BD5189A1EAA68E2C8310EA1808D8C499A5E104DD3456A16956C9F5F1B4C83AAFAE2376F1979AC81F165F3D53BADBFABCFDD76E4787DD7E9ADB7ABDDEBE7F67DE41C0F6FB70FCE1843595D86E54BB9FFE3D4D96D35D4E2C8F9381CF64384CEE773BF2AF7FDD56E834C403B9C267E2ABF8EEB0A0391922C3128E9C7733A8B8A088346859291901267248D6E5CD72324E4A12229F5007564A03346A3829384612AA621A00E614599734C5BC11420C5A2E079326112BBE6E89E1D0A922C4845D6482E205B1A3AF4B19205C9B3B8A0B1C8FD42C4B43668C1E85C60D7F36A4643C3CC89F4FE1DA2D5ACE90EB568F08BEE1BA496297EAE030601A12A141C134016414232C5A416195664926BEB95B062FD1F2951EEA0C866ADA1F9A43448F4EBE54049400D0631796454D966B630E49F2E2D03F3902D0C42B56B9A73AA914E476CF9EFEC56824C48358D49A0DDF68A09375733992F38BBCE5957206418D89F60FA05D70A147B52A1AFC3A8BB4AC8AE17EAAE1C6AD717FF00504B010213001400000008009B9B5752BE5CF063BB79000090360500160000000000000000000000000000000000534150452D3030314141494E4954415043312E583032504B010213001400000008009B9B5752DFC61169BF280000D94102000800000000000000000000000000EF790000554C4F4732315F31504B010213001400000008009B9B5752E5A69501B8010000460D00000800000000000000000000000000D4A20000414C4F4732313038504B010213001400000008009B9B57520175BE318D010000A60A00000C00000000000000000000000000B2A40000534C4F47323130382E583032504B010213001400000008009B9B5752EC9F64BEEC00000062010000090000000000000000000000000069A6000044445052482E786D6C504B010213001400000008009B9B57523F481178EF0600000142000009000000000000000000000000007CA7000044445052532E786D6C504B010213001400000008009B9B5752A85568E016010000D10100000E0000000000000000000000000092AE0000545244454C49564552592E786D6C504B010213001400000008009B9B5752EAC1A9D03C010000A10600001100000000000000000000000000D4AF0000545244454C5643484B534554442E786D6C504B010213001400000008009B9B5752250DBBB5BA0000002B0100000E000000000000000000000000003FB10000545244454C5652454C562E786D6C504B010213001400000008009B9B575225D773F5CA0000002D0100000E0000000000000000000000000025B20000545244454C5652454C562E786D6C504B010213001400000008009B9B575239AE7843F70000008A0200000E000000000000000000000000001BB30000545244454C565441534B2E786D6C504B010213001400000008009B9B57520CB80CB33D0100002A0200000E000000000000000000000000003EB40000545244454C56554E49542E786D6C504B010213001400000008009B9B5752B152EFE1DC010000830E00000E00000000000000000000000000A7B500005452444C5643484B5F482E786D6C504B010213001400000008009B9B57529E26DAC047010000980200000E00000000000000000000000000AFB700005452444C5643484B5F482E786D6C504B010213001400000008009B9B5752EA7417389C010000460500000E0000000000000000000000000022B900005452444C5643484B5F4B2E786D6C504B010213001400000008009B9B575252F5D8F07F010000EF0200000E00000000000000000000000000EABA00005452444C5643484B5F4B2E786D6C504B05060000000010001000B303000095BC00000000",
				"mimetype" : "application/x-zip-compressed"
			}
		]
	}
}`,
	StatusCode: 200,
}

var buildGetTask11ResultMedia = MockData{
	Method:     `GET`,
	Url:        `/sap/opu/odata/BUILD/CORE_SRV/results(build_id='AKO22FYOFYPOXHOBVKXUTX3A3Q',task_id=11,name='SAR_XML')/$value`,
	Body:       ``,
	StatusCode: 200,
}

var buildGetValues = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/values`,
	Body: `{
		"d": {
			"results": [
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "PHASE",
					"value": "AUNIT"
				},
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "PACKAGES",
					"value": "/BUILD/AUNIT_DUMMY_TESTS"
				},
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "MyId1",
					"value": "AunitValue1"
				},
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "MyId2",
					"value": "AunitValue2"
				},
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "BUILD_FRAMEWORK_MODE",
					"value": "P"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var buildGetValuesWithClient = MockData{
	Method: `GET`,
	Url:    `/sap/opu/odata/BUILD/CORE_SRV/builds('AKO22FYOFYPOXHOBVKXUTX3A3Q')/values?sap-client=001`,
	Body: `{
		"d": {
			"results": [
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "PHASE",
					"value": "AUNIT"
				},
				{
					"build_id": "AKO22FYOFYPOXHOBVKXUTX3A3Q",
					"value_id": "SUN",
					"value": "SUMMER"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var template = MockData{
	Method:     `GET`,
	Url:        ``,
	Body:       ``,
	StatusCode: 200,
}
