//go:build unit

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
)

type protecodeTestUtilsBundle struct {
	*mock.FilesMock
	*mock.DownloadMock
}

func TestRunProtecodeScan(t *testing.T) {
	requestURI := ""

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		var b bytes.Buffer

		if requestURI == "/api/product/4486/" || requestURI == "/api/product/4711/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := os.ReadFile(violations)
			require.NoErrorf(t, err, "failed reading %v", violations)
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			err = json.Unmarshal(byteContent, &response)

			_ = json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/fetch/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := os.ReadFile(violations)
			require.NoErrorf(t, err, "failed reading %v", violations)
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4486, ReportURL: requestURI}}
			err = json.Unmarshal(byteContent, &response)

			_ = json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/product/4486/pdf-report" {

		} else if requestURI == "/api/upload/t.tar" {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4486, ReportURL: requestURI}}

			var b bytes.Buffer
			_ = json.NewEncoder(&b).Encode(&response)
			if _, err := rw.Write([]byte(b.Bytes())); err != nil {
				t.Fail()
			}
		} else {
			response := protecode.Result{ProductID: 4486, ReportURL: requestURI}
			_ = json.NewEncoder(&b).Encode(&response)
		}

		if _, err := rw.Write([]byte(b.Bytes())); err != nil {
			t.Fail()
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.Options{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	influx := protecodeExecuteScanInflux{}

	ttt := []struct {
		name string
		opts protecodeExecuteScanOptions
	}{
		{
			name: "With tar as scan image",
			opts: protecodeExecuteScanOptions{
				ServerURL:      server.URL,
				TimeoutMinutes: "1",
				VerifyOnly:     false,
				CleanupMode:    "none",
				Group:          "13",
				FetchURL:       "/api/fetch/",
				ExcludeCVEs:    "CVE-2018-1, CVE-2017-1000382",
				ReportFileName: "./cache/report-file.txt",
			},
		},
		{
			name: "Without tar as scan image",
			opts: protecodeExecuteScanOptions{
				ServerURL:      server.URL,
				ScanImage:      "t",
				TimeoutMinutes: "1",
				VerifyOnly:     false,
				CleanupMode:    "none",
				Group:          "13",
				ExcludeCVEs:    "CVE-2018-1, CVE-2017-1000382",
				ReportFileName: "./cache/report-file.txt",
			},
		},
	}

	for _, test := range ttt {
		t.Run(test.name, func(t *testing.T) {
			files := mock.FilesMock{}

			httpClient := piperHttp.Client{}
			httpClient.SetFileUtils(&files)

			pc.SetHttpClient(&httpClient)

			docker := mock.DownloadMock{}
			docker.Stub = func(imageRef string, dest string) (v1.Image, error) {
				files.AddFile(dest, []byte(""))
				return &fake.FakeImage{}, nil
			}

			utils := protecodeTestUtilsBundle{DownloadMock: &docker, FilesMock: &files}

			cacheDir, _ := files.TempDir("", "protecode-")
			files.AddFile(filepath.Join(cacheDir, "t.tar"), []byte(""))

			err := runProtecodeScan(&test.opts, &influx, pc, utils, cacheDir)

			if assert.NoError(t, err) {
				if protecodeExecuteJsonExists, err := files.FileExists("protecodeExecuteScan.json"); assert.NoError(t, err) {
					assert.True(t, protecodeExecuteJsonExists, "protecodeExecuteScan.json expected")
				}

				if protecodeVulnsJsonExists, err := files.FileExists("protecodescan_vulns.json"); assert.NoError(t, err) {
					assert.True(t, protecodeVulnsJsonExists, "protecodescan_vulns.json expected")
				}

				if userSpecifiedReportExists, err := files.FileExists(test.opts.ReportFileName); assert.NoError(t, err) {
					assert.True(t, userSpecifiedReportExists, "%s must exist", test.opts.ReportFileName)
				}
			}

			if cacheExists, err := files.DirExists(cacheDir); assert.NoError(t, err) {
				assert.True(t, cacheExists, "Whoever calls runProtecodeScan is responsible to cleanup the cache.")
			}
		})
	}
}

func TestHandleArtifactVersion(t *testing.T) {
	cases := []struct {
		version string
		want    string
	}{
		{"1.0.0-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e", "1"},
		{"2.20.20-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e", "2"},
		{"3.20.20-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e", "3"},
		{"4.20.20-20200131085038", "4"},
		{"5.20.20-20200131085038+", "5"},
		{"6.00", "6.00"},
		{"7.20.20", "7.20.20"},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := handleArtifactVersion(c.version)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestCreateProtecodeClient(t *testing.T) {
	cases := []struct {
		timeout string
	}{
		{""},
		{"1"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("With timeout: %s", c.timeout), func(t *testing.T) {
			config := protecodeExecuteScanOptions{TimeoutMinutes: c.timeout}

			client := createProtecodeClient(&config)
			assert.NotNil(t, client, "client should not be empty")
		})
	}
}

func TestUploadScanOrDeclareFetch(t *testing.T) {
	// init
	testFile, err := os.CreateTemp("", "testFileUpload")
	require.NoError(t, err)
	defer os.RemoveAll(testFile.Name()) // clean up
	fileName := filepath.Base(testFile.Name())
	path := strings.ReplaceAll(testFile.Name(), fileName, "")

	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI

		if requestURI == "/api/fetch/" {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			var b bytes.Buffer
			_ = json.NewEncoder(&b).Encode(&response)
			if _, err := rw.Write([]byte(b.Bytes())); err != nil {
				t.Fail()
			}
		}
		if requestURI == fmt.Sprintf("/api/upload/%v", fileName) || requestURI == fmt.Sprintf("/api/upload/PR_4711_%v", fileName) {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}

			var b bytes.Buffer
			_ = json.NewEncoder(&b).Encode(&response)
			if _, err := rw.Write([]byte(b.Bytes())); err != nil {
				t.Fail()
			}
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.Options{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	cases := []struct {
		reuse    bool
		clean    string
		group    string
		fetchURL string
		filePath string
		prName   string
		prID     int
		want     int
	}{
		{false, "test", "group1", "/api/fetch/", "", "", -1, 4711},
		{true, "test", "group1", "/api/fetch/", "", "", -1, 4711},
		{false, "test", "group1", "/api/fetch/", "", "", 4711, 4711},
		{false, "test", "group1", "/api/fetch/", "", "", 0, 4711},

		{false, "test", "group1", "", path, "", -1, 4711},
		{true, "test", "group1", "", path, "", -1, 4711},
		{false, "test", "group1", "", path, "PR_4711", 4711, 4711},
		{false, "test", "group1", "", path, "", 0, 4711},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			utils := protecodeTestUtilsBundle{
				FilesMock:    &mock.FilesMock{},
				DownloadMock: &mock.DownloadMock{},
			}

			// test
			config := protecodeExecuteScanOptions{VerifyOnly: c.reuse, CleanupMode: c.clean, Group: c.group, FetchURL: c.fetchURL, FilePath: c.filePath}
			// got := uploadScanOrDeclareFetch(config, 0, pc, fileName)
			got := uploadScanOrDeclareFetch(utils, config, c.prID, pc, fileName)
			// assert
			assert.Equal(t, c.want, got)
		})
	}
}

func TestExecuteProtecodeScan(t *testing.T) {
	testDataFile := filepath.Join("testdata", "TestProtecode", "protecode_result_violations.json")
	violationsAbsPath, err := filepath.Abs(testDataFile)
	require.NoErrorf(t, err, "failed to obtain absolute path to test data with violations: %v", err)

	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		var b bytes.Buffer

		if requestURI == "/api/product/4711/" {
			byteContent, err := os.ReadFile(violationsAbsPath)
			require.NoErrorf(t, err, "failed reading %v", violationsAbsPath)
			response := protecode.ResultData{}
			err = json.Unmarshal(byteContent, &response)

			_ = json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/product/4711/pdf-report" {

		} else {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			_ = json.NewEncoder(&b).Encode(&response)
		}

		if _, err := rw.Write([]byte(b.Bytes())); err != nil {
			t.Fail()
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.Options{ServerURL: server.URL, Duration: time.Minute * 3}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	cases := []struct {
		reuse    bool
		clean    string
		group    string
		fetchURL string
		want     int
	}{
		{false, "binary", "group1", "/api/fetch/", 4711},
	}

	resetDir, err := os.Getwd()
	require.NoErrorf(t, err, "Failed to get current directory: %v", err)
	defer func() { _ = os.Chdir(resetDir) }()

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			utils := protecodeTestUtilsBundle{
				FilesMock:    &mock.FilesMock{},
				DownloadMock: &mock.DownloadMock{},
			}
			config := protecodeExecuteScanOptions{VerifyOnly: c.reuse, CleanupMode: c.clean, Group: c.group, FetchURL: c.fetchURL, TimeoutMinutes: "3", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}
			influxData := &protecodeExecuteScanInflux{}
			// test
			executeProtecodeScan(influxData, pc, &config, "dummy", utils)
			// assert
			assert.Equal(t, 1125, influxData.protecode_data.fields.historical_vulnerabilities)
			assert.Equal(t, 0, influxData.protecode_data.fields.triaged_vulnerabilities)
			assert.Equal(t, 1, influxData.protecode_data.fields.excluded_vulnerabilities)
			assert.Equal(t, 129, influxData.protecode_data.fields.major_vulnerabilities)
			assert.Equal(t, 226, influxData.protecode_data.fields.vulnerabilities)
		})
	}
}

func TestCorrectDockerConfigEnvVar(t *testing.T) {
	utils := protecodeTestUtilsBundle{
		FilesMock:    &mock.FilesMock{},
		DownloadMock: &mock.DownloadMock{},
	}

	t.Run("with credentials", func(t *testing.T) {
		// init
		testDirectory := t.TempDir()

		dockerConfigDir := filepath.Join(testDirectory, "myConfig")
		if err := os.Mkdir(dockerConfigDir, 0755); err != nil {
			t.Fail()
		}
		require.DirExists(t, dockerConfigDir)

		dockerConfigFile := filepath.Join(dockerConfigDir, "docker.json")
		file, _ := os.Create(dockerConfigFile)
		defer file.Close()
		require.FileExists(t, dockerConfigFile)

		resetValue := os.Getenv("DOCKER_CONFIG")
		defer os.Setenv("DOCKER_CONFIG", resetValue)
		// test
		correctDockerConfigEnvVar(&protecodeExecuteScanOptions{DockerConfigJSON: dockerConfigFile}, utils)
		// assert
		absolutePath, _ := filepath.Abs(dockerConfigDir)
		assert.Equal(t, absolutePath, os.Getenv("DOCKER_CONFIG"))
	})
	t.Run("without credentials", func(t *testing.T) {
		// init
		resetValue := os.Getenv("DOCKER_CONFIG")
		defer os.Setenv("DOCKER_CONFIG", resetValue)
		// test
		correctDockerConfigEnvVar(&protecodeExecuteScanOptions{}, utils)
		// assert
		assert.Equal(t, resetValue, os.Getenv("DOCKER_CONFIG"))
	})
}
