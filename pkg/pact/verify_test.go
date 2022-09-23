package pact

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecPactVerify(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}
		err := vConfig.ExecPactVerify()
		assert.NoError(t, err)
	})

	t.Run("failure - save report", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.FileWriteError = fmt.Errorf("write failed")
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}
		err := vConfig.ExecPactVerify()
		assert.Contains(t, fmt.Sprint(err), "error saving report")
	})

	t.Run("failure - tests failed", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.ExitCode = 1
		mockUtils.ShouldFailOnCommand = map[string]error{"swagger-mock-validator": fmt.Errorf("test error")}
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}

		err := vConfig.ExecPactVerify()
		assert.Contains(t, fmt.Sprint(err), "contract tests failed, http: Failed, asynch: Failed")
	})

	t.Run("failure - tests failed - asynch only", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}

		err := vConfig.ExecPactVerify()
		assert.Contains(t, fmt.Sprint(err), "contract tests failed") 
		assert.Contains(t, fmt.Sprint(err), "asynch: Failed")
	})

	t.Run("failure - pact verification - http", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"https://the.url/pacts/provider/testProvider-http/latest/main": fmt.Errorf("send error")}
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PactBrokerBaseURL: "the.url",
			GitTargetBranch: "main",
			Provider: "testProvider",

		}
		err := vConfig.ExecPactVerify()
		assert.Contains(t, fmt.Sprint(err), "contract tests validation failed, http: send error")
	})

	t.Run("failure - pact verification - async", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		asynchFile := "asyncapidoc.json"
		mockUtils.AddFile(asynchFile, []byte(`{}`))
		mockUtils.AddFile("swagger.json", []byte(`{}`))
		mockUtils.ShouldFailOnCommand = map[string]error{"async-api-validator": fmt.Errorf("exec failure")}

		vConfig := VerifyConfig{
			Utils: mockUtils,
			Provider: "testProvider",
			PathToAsyncFile: asynchFile,
			PathToSwaggerFile: "swagger.json",
		}
		err := vConfig.ExecPactVerify()
		assert.Contains(t, fmt.Sprint(err), "contract tests validation failed")
	})
}

func TestVerifyHttp(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		mockUtils.AddFile("swagger.json", []byte(`{}`))
		vConfig := VerifyConfig{
			Utils: mockUtils,
			GitTargetBranch: "main",
			PathToSwaggerFile: "swagger.json",
			PactBrokerBaseURL: "pact.broker.url",
			PactBrokerUsername: "testUser",
			PactBrokerPassword: "testPassword",
			Provider: "testProvider",
		}
		exitCode, noTests, err := vConfig.verifyHTTP()
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, 1, noTests)
		assert.Equal(t, "swagger-mock-validator", mockUtils.Calls[0].Exec)
		expectedParams := []string{
			vConfig.PathToSwaggerFile,
			"https://pact.broker.url",
			"--provider=testProvider-http",
			"--tag=main",
			"--user=testUser:testPassword",
		}
		assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
	})

	t.Run("success - not found", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseStatusCode = http.StatusNotFound
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}
		exitCode, noTests, err := vConfig.verifyHTTP()
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, 0, noTests)
	})

	t.Run("success - no tests", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}
		exitCode, noTests, err := vConfig.verifyHTTP()
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, 0, noTests)
	})

	t.Run("failure - download error", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"https://the.url/pacts/provider/theProvider-http/latest/main": fmt.Errorf("send failure")}
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PactBrokerBaseURL: "the.url",
			Provider: "theProvider",
			GitTargetBranch: "main",
		}
		_, _, err := vConfig.verifyHTTP()
		assert.EqualError(t, err, "send failure")
	})

	t.Run("exit 1 - no swagger file", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToSwaggerFile: "swagger.json",
		}
		exitCode, _, err := vConfig.verifyHTTP()
		assert.NoError(t, err)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("exit 1 - swagger file find error", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.FileExistsErrors = map[string]error{"swagger.json": fmt.Errorf("not found")}
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToSwaggerFile: "swagger.json",
		}
		exitCode, _, err := vConfig.verifyHTTP()
		assert.NoError(t, err)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("failure - lookPath", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.LookPathError = fmt.Errorf("lookPath error")
		mockUtils.AddFile("swagger.json", []byte(`{}`))
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToSwaggerFile: "swagger.json",
		}
		exitCode, _, err := vConfig.verifyHTTP()
		assert.EqualError(t, err, "lookPath error")
		assert.Equal(t, 1, exitCode)
	})

	t.Run("failure - pact execution", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.ShouldFailOnCommand = map[string]error{"swagger-mock-validator": fmt.Errorf("test error")}
		mockUtils.ExitCode = 1
		mockUtils.AddFile("swagger.json", []byte(`{}`))
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToSwaggerFile: "swagger.json",
		}
		exitCode, _, err := vConfig.verifyHTTP()
		assert.EqualError(t, err, "test error")
		assert.Equal(t, 1, exitCode)
	})
}

func TestVerifyAsync(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		mockUtils.AddFile("path/to/asynch-file.json", []byte(`{}`))
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToAsyncFile: "path/to/asynch-file.json",
		}
		exitCode, noTests, err := vConfig.verifyAsync()
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, 1, noTests)
		assert.Equal(t, "async-api-validator", mockUtils.Calls[0].Exec)
		expectedParams := []string{
			"--pathToPactFolder=./async_verify_pacts/",
			"--pathToAsyncFile=path/to/asynch-file.json",

		}
		assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)

	})

	t.Run("success - not found", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseStatusCode = http.StatusNotFound
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}
		exitCode, noTests, err := vConfig.verifyAsync()
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, 0, noTests)
	})

	t.Run("success - no tests", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
		}
		exitCode, noTests, err := vConfig.verifyAsync()
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, 0, noTests)
	})

	t.Run("failure - download error", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpSendErrors = map[string]error{"https://the.url/pacts/provider/testProvider-async/latest/main": fmt.Errorf("send failure")}
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PactBrokerBaseURL: "the.url",
			Provider: "testProvider",
			GitTargetBranch: "main",
		}
		_, _, err := vConfig.verifyAsync()
		assert.EqualError(t, err, "send failure")
	})

	t.Run("exit 1 - no asynch file", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToAsyncFile: "path/to/asynch-file.json",
		}
		exitCode, _, err := vConfig.verifyAsync()
		assert.NoError(t, err)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("exit 1 - swagger file find error", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.FileExistsErrors = map[string]error{"path/to/asynch-file.json": fmt.Errorf("not found")}
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToAsyncFile: "path/to/asynch-file.json",
		}
		exitCode, _, err := vConfig.verifyAsync()
		assert.NoError(t, err)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("failure - lookPath", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.LookPathError = fmt.Errorf("lookPath error")
		mockUtils.AddFile("path/to/asynch-file.json", []byte(`{}`))
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToAsyncFile: "path/to/asynch-file.json",
		}
		exitCode, _, err := vConfig.verifyAsync()
		assert.EqualError(t, err, "lookPath error")
		assert.Equal(t, 1, exitCode)
	})

	t.Run("failure - pact execution", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.ShouldFailOnCommand = map[string]error{"async-api-validator": fmt.Errorf("test error")}
		mockUtils.ExitCode = 1
		mockUtils.AddFile("path/to/asynch-file.json", []byte(`{}`))
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			PathToAsyncFile: "path/to/asynch-file.json",
		}
		exitCode, _, err := vConfig.verifyAsync()
		assert.EqualError(t, err, "test error")
		assert.Equal(t, 1, exitCode)
	})
}

func TestDownloadContractsToDisk(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
					{HRef: "https://link.2", Title: "contract 2", Name: "2"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		vConfig := VerifyConfig{
			Utils: mockUtils,
			Provider: "testProvider",
		}

		numTests, err := vConfig.downloadContractsToDisk()
		assert.NoError(t, err)
		assert.Equal(t, 2, numTests)
		assert.True(t, mockUtils.HasWrittenFile("./async_verify_pacts/1-testProvider-async.json"), "First PACT missing")
		assert.True(t, mockUtils.HasWrittenFile("./async_verify_pacts/2-testProvider-async.json"), "Second PACT missing")
	})

	t.Run("failure - get pacts", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.httpResponseStatusCode = http.StatusNotFound
		vConfig := VerifyConfig{
			Utils: mockUtils,
			Provider: "testProvider",
		}

		_, err := vConfig.downloadContractsToDisk()
		assert.EqualError(t, err, "404: no consumer tests found for provider")
	})

	t.Run("failure - ensure directory", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.DirCreateErrors = map[string]error{"./async_verify_pacts/": fmt.Errorf("create failure")}
		vConfig := VerifyConfig{
			Utils: mockUtils,
			Provider: "testProvider",
		}

		_, err := vConfig.downloadContractsToDisk()
		assert.Contains(t, fmt.Sprint(err), "failed to ensure that directory is existing")
	})

	t.Run("failure - download contract", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		mockUtils.httpSendErrors = map[string]error{"https://link.1": fmt.Errorf("send failure")}
		vConfig := VerifyConfig{
			Utils: mockUtils,
			Provider: "testProvider",
		}

		_, err := vConfig.downloadContractsToDisk()
		assert.Contains(t, fmt.Sprint(err), "send failure")
	})

	t.Run("failure - write pact", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		testLinks := LatestPactsForProviderTagResp{
			Links: Links{
				PBPacts: []Link{
					{HRef: "https://link.1", Title: "contract 1", Name: "1"},
				},
			},
		}
		mockUtils.pbLinks = testLinks
		mockUtils.FileWriteError = fmt.Errorf("write failed")
		vConfig := VerifyConfig{
			Utils: mockUtils,
			Provider: "testProvider",
		}

		_, err := vConfig.downloadContractsToDisk()
		assert.Contains(t, fmt.Sprint(err), "write failed")
	})
}

func TestEnforce(t *testing.T) {
	t.Run("success", func(t *testing.T){
		vConfig := VerifyConfig{
			EnforceAsyncAPIValidation: true,
			EnforceOpenAPIValidation: true,
		}
		err := vConfig.Enforce(0, 0)
		assert.NoError(t, err)
	})

	t.Run("success - not enabled", func(t *testing.T){
		vConfig := VerifyConfig{}
		err := vConfig.Enforce(1, 0)
		assert.NoError(t, err)
	})

	t.Run("failure - openapi validation", func(t *testing.T){
		vConfig := VerifyConfig{
			EnforceOpenAPIValidation: true,
		}
		err := vConfig.Enforce(1, 0)
		assert.EqualError(t, err, fmt.Sprint(ErrEnforcement))
	})

	t.Run("failure - asyncapi validation", func(t *testing.T){
		vConfig := VerifyConfig{
			EnforceAsyncAPIValidation: true,
		}
		err := vConfig.Enforce(0, 1)
		assert.EqualError(t, err, fmt.Sprint(ErrEnforcement))
	})
}

func TestCheckThreshold(t *testing.T) {
	
}
