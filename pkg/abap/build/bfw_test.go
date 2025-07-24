//go:build unit
// +build unit

package build

import (
	"encoding/json"
	"path"
	"path/filepath"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func testSetup(client piperhttp.Sender, buildID string) Build {
	conn := new(Connector)
	conn.Client = client
	conn.DownloadClient = &DownloadClientMock{}
	conn.Header = make(map[string][]string)
	b := Build{
		Connector: *conn,
		BuildID:   buildID,
	}
	return b
}

func TestStart(t *testing.T) {
	t.Run("Run start", func(t *testing.T) {
		client := &ClMock{
			Token: "MyToken",
		}
		b := testSetup(client, "")
		inputValues := Values{
			Values: []Value{
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
		err := b.Start("test", inputValues)
		assert.NoError(t, err)
		assert.Equal(t, Accepted, b.RunState)
	})
}

func TestStartValueGeneration(t *testing.T) {
	myValue := string(`{ "ARES_EC_ATTRIBUTES": [ { "ATTRIBUTE": "A", "VALUE": "B" } ] }`)

	inputForPost := InputForPost{
		Phase:  "HUGO",
		Values: []Value{{ValueID: "myJson", Value: myValue}},
	}

	importBody, err := json.Marshal(inputForPost)
	assert.NoError(t, err)
	assert.Equal(t, `{"phase":"HUGO","values":[{"value_id":"myJson","value":"{ \"ARES_EC_ATTRIBUTES\": [ { \"ATTRIBUTE\": \"A\", \"VALUE\": \"B\" } ] }"}]}`, string(importBody))
}

func TestGet(t *testing.T) {
	t.Run("Run Get", func(t *testing.T) {
		b := testSetup(&ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		err := b.Get()
		assert.NoError(t, err)
		assert.Equal(t, Finished, b.RunState)
		assert.Equal(t, 0, len(b.Tasks))
	})
}

func TestGetTasks(t *testing.T) {
	t.Run("Run getTasks", func(t *testing.T) {
		b := testSetup(&ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
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
		b := testSetup(&ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
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
		b := testSetup(&ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		assert.Equal(t, 0, len(b.Values))
		err := b.GetValues()
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
		b := testSetup(&ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		err := b.GetResults()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(b.Tasks[0].Results))
		assert.Equal(t, 2, len(b.Tasks[1].Results))
		assert.Equal(t, "image/jpeg", b.Tasks[1].Results[0].Mimetype)
		assert.Equal(t, "application/octet-stream", b.Tasks[1].Results[1].Mimetype)

		_, err = b.GetResult("does_not_exist")
		assert.Error(t, err)
		r, err := b.GetResult("SAR_XML")
		assert.Equal(t, "application/octet-stream", r.Mimetype)
		assert.NoError(t, err)
	})
}

func TestPoll(t *testing.T) {
	//arrange global
	build := new(Build)
	conn := new(Connector)
	conn.MaxRuntime = time.Duration(1 * time.Second)
	conn.PollingInterval = time.Duration(1 * time.Microsecond)
	conn.Baseurl = "/sap/opu/odata/BUILD/CORE_SRV"
	t.Run("Normal Poll", func(t *testing.T) {
		//arrange
		build.BuildID = "AKO22FYOFYPOXHOBVKXUTX3A3Q"
		mc := NewMockClient()
		mc.AddData(buildGet1)
		mc.AddData(buildGet2)
		conn.Client = &mc
		build.Connector = *conn
		//act
		err := build.Poll()
		//assert
		assert.NoError(t, err)
	})
	t.Run("Poll runstate failed", func(t *testing.T) {
		//arrange
		build.BuildID = "AKO22FYOFYPOXHOBVKXUTX3A3Q"
		mc := NewMockClient()
		mc.AddData(buildGet1)
		mc.AddData(buildGetRunStateFailed)
		conn.Client = &mc
		build.Connector = *conn
		//act
		err := build.Poll()
		//assert
		assert.NoError(t, err)
	})
	t.Run("Poll timeout", func(t *testing.T) {
		//arrange
		build.BuildID = "AKO22FYOFYPOXHOBVKXUTX3A3Q"
		conn.MaxRuntime = time.Duration(1 * time.Microsecond)
		conn.PollingInterval = time.Duration(1 * time.Microsecond)
		mc := NewMockClient()
		mc.AddData(buildGet1)
		mc.AddData(buildGet1)
		mc.AddData(buildGet2)
		conn.Client = &mc
		build.Connector = *conn
		//act
		err := build.Poll()
		//assert
		assert.Error(t, err)
	})
}

func TestEvaluteIfBuildSuccessful(t *testing.T) {
	//arrange global
	build := new(Build)
	treatWarningsAsError := false
	t.Run("No error", func(t *testing.T) {
		//arrange
		build.RunState = Finished
		build.ResultState = Successful
		//act
		err := build.EvaluteIfBuildSuccessful(treatWarningsAsError)
		//assert
		assert.NoError(t, err)
	})
	t.Run("RunState failed => Error", func(t *testing.T) {
		//arrange
		build.RunState = Failed
		//act
		err := build.EvaluteIfBuildSuccessful(treatWarningsAsError)
		//assert
		assert.Error(t, err)
	})
	t.Run("ResultState aborted => Error", func(t *testing.T) {
		//arrange
		build.RunState = Finished
		build.ResultState = Aborted
		//act
		err := build.EvaluteIfBuildSuccessful(treatWarningsAsError)
		//assert
		assert.Error(t, err)
	})
	t.Run("ResultState erroneous => Error", func(t *testing.T) {
		//arrange
		build.RunState = Finished
		build.ResultState = Erroneous
		//act
		err := build.EvaluteIfBuildSuccessful(treatWarningsAsError)
		//assert
		assert.Error(t, err)
	})
	t.Run("ResultState warning, treatWarningsAsError false => No error", func(t *testing.T) {
		//arrange
		build.RunState = Finished
		build.ResultState = Warning
		//act
		err := build.EvaluteIfBuildSuccessful(treatWarningsAsError)
		//assert
		assert.NoError(t, err)
	})
	t.Run("ResultState warning, treatWarningsAsError true => error", func(t *testing.T) {
		//arrange
		build.RunState = Finished
		build.ResultState = Warning
		treatWarningsAsError = true
		//act
		err := build.EvaluteIfBuildSuccessful(treatWarningsAsError)
		//assert
		assert.Error(t, err)
	})
}

func TestDownloadWithFilenamePrefixAndTargetDirectory(t *testing.T) {
	//arrange global
	result := new(Result)
	result.BuildID = "123456789"
	result.TaskID = 1
	result.Name = "MyFile"
	conn := new(Connector)
	conn.DownloadClient = &DownloadClientMock{}
	result.connector = *conn
	t.Run("Download without extension", func(t *testing.T) {
		//arrange
		basePath := ""
		filenamePrefix := ""
		//act
		err := result.DownloadWithFilenamePrefixAndTargetDirectory(basePath, filenamePrefix)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "MyFile", result.SavedFilename)
		assert.Equal(t, "MyFile", result.DownloadPath)
	})
	t.Run("Download with extensions", func(t *testing.T) {
		//arrange
		basePath := "MyDir"
		filenamePrefix := "SuperFile_"
		//act
		err := result.DownloadWithFilenamePrefixAndTargetDirectory(basePath, filenamePrefix)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "SuperFile_MyFile", result.SavedFilename)
		downloadPath := filepath.Join(path.Base(basePath), path.Base("SuperFile_MyFile"))
		assert.Equal(t, downloadPath, result.DownloadPath)
	})
	t.Run("Download with parameter", func(t *testing.T) {
		//arrange
		basePath := "{BuildID}"
		filenamePrefix := "{taskid}"
		//act
		err := result.DownloadWithFilenamePrefixAndTargetDirectory(basePath, filenamePrefix)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "1MyFile", result.SavedFilename)
		downloadPath := filepath.Join(path.Base("123456789"), path.Base("1MyFile"))
		assert.Equal(t, downloadPath, result.DownloadPath)
	})
}

func TestDownloadAllResults(t *testing.T) {
	//arrange global
	build := GetMockBuildTestDownloadPublish()
	t.Run("Download without extension", func(t *testing.T) {
		//arrange
		basePath := ""
		filenamePrefix := ""
		//act
		err := build.DownloadAllResults(basePath, filenamePrefix)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "", build.Tasks[0].Results[0].SavedFilename)
		assert.Equal(t, "", build.Tasks[0].Results[0].DownloadPath)

		assert.Equal(t, "File1", build.Tasks[1].Results[0].SavedFilename)
		assert.Equal(t, "File1", build.Tasks[1].Results[0].DownloadPath)

		assert.Equal(t, "File2", build.Tasks[1].Results[1].SavedFilename)
		assert.Equal(t, "File2", build.Tasks[1].Results[1].DownloadPath)
	})
	t.Run("Download with extension", func(t *testing.T) {
		//arrange
		basePath := ""
		filenamePrefix := "SuperFile_"
		//act
		err := build.DownloadAllResults(basePath, filenamePrefix)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "", build.Tasks[0].Results[0].SavedFilename)
		assert.Equal(t, "", build.Tasks[0].Results[0].DownloadPath)

		assert.Equal(t, "SuperFile_File1", build.Tasks[1].Results[0].SavedFilename)
		assert.Equal(t, "SuperFile_File1", build.Tasks[1].Results[0].DownloadPath)

		assert.Equal(t, "SuperFile_File2", build.Tasks[1].Results[1].SavedFilename)
		assert.Equal(t, "SuperFile_File2", build.Tasks[1].Results[1].DownloadPath)
	})
}

func TestDownloadResults(t *testing.T) {
	//arrange global
	build := GetMockBuildTestDownloadPublish()
	t.Run("Download existing", func(t *testing.T) {
		//arrange
		basePath := ""
		filenamePrefix := ""
		filenames := []string{"File1", "File3"}
		//act
		err := build.DownloadResults(filenames, basePath, filenamePrefix)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "", build.Tasks[0].Results[0].SavedFilename)
		assert.Equal(t, "", build.Tasks[0].Results[0].DownloadPath)

		assert.Equal(t, "File1", build.Tasks[1].Results[0].SavedFilename)
		assert.Equal(t, "File1", build.Tasks[1].Results[0].DownloadPath)

		assert.Equal(t, "", build.Tasks[1].Results[1].SavedFilename)
		assert.Equal(t, "", build.Tasks[1].Results[1].DownloadPath)

		assert.Equal(t, "File3", build.Tasks[1].Results[2].SavedFilename)
		assert.Equal(t, "File3", build.Tasks[1].Results[2].DownloadPath)
	})
	t.Run("Try to download non existing", func(t *testing.T) {
		//arrange
		basePath := ""
		filenamePrefix := ""
		filenames := []string{"File1", "File4"}
		//act
		err := build.DownloadResults(filenames, basePath, filenamePrefix)
		//assert
		assert.Error(t, err)
		assert.Equal(t, "", build.Tasks[0].Results[0].SavedFilename)
		assert.Equal(t, "", build.Tasks[0].Results[0].DownloadPath)

		assert.Equal(t, "File1", build.Tasks[1].Results[0].SavedFilename)
		assert.Equal(t, "File1", build.Tasks[1].Results[0].DownloadPath)

		assert.Equal(t, "", build.Tasks[1].Results[1].SavedFilename)
		assert.Equal(t, "", build.Tasks[1].Results[1].DownloadPath)
	})
}

func TestPublishAllDownloadedResults(t *testing.T) {
	t.Run("Something was downloaded", func(t *testing.T) {
		//arrange
		build := GetMockBuildTestDownloadPublish()
		files := mock.FilesMock{}
		build.Tasks[1].Results[0].SavedFilename = "File1"
		build.Tasks[1].Results[0].DownloadPath = "Dir1/File1"
		build.Tasks[1].Results[2].SavedFilename = "File3"
		build.Tasks[1].Results[2].DownloadPath = "File3"
		//act
		build.PublishAllDownloadedResults("MyStep", &files)
		//assert
		assert.True(t, files.HasFile("/MyStep_reports.json"))
		assert.True(t, files.HasFile("/MyStep_links.json"))
	})
	t.Run("Nothing was downloaded", func(t *testing.T) {
		//arrange
		build := GetMockBuildTestDownloadPublish()
		files := mock.FilesMock{}
		//act
		build.PublishAllDownloadedResults("MyStep", &files)
		//assert
		assert.False(t, files.HasFile("/MyStep_reports.json"))
		assert.False(t, files.HasFile("/MyStep_links.json"))
	})
}

func TestPublishDownloadedResults(t *testing.T) {
	filenames := []string{"File1", "File3"}
	t.Run("Publish downloaded files", func(t *testing.T) {
		//arrange
		build := GetMockBuildTestDownloadPublish()
		files := mock.FilesMock{}
		build.Tasks[1].Results[0].SavedFilename = "SuperFile_File1"
		build.Tasks[1].Results[0].DownloadPath = "Dir1/SuperFile_File1"
		build.Tasks[1].Results[2].SavedFilename = "File3"
		build.Tasks[1].Results[2].DownloadPath = "File3"
		//act
		err := build.PublishDownloadedResults("MyStep", filenames, &files)
		//assert
		assert.NoError(t, err)

		assert.True(t, files.HasFile("/MyStep_reports.json"))
		assert.True(t, files.HasFile("/MyStep_links.json"))

	})
	t.Run("Try to publish file which was not downloaded", func(t *testing.T) {
		//arrange
		build := GetMockBuildTestDownloadPublish()
		files := mock.FilesMock{}
		build.Tasks[1].Results[0].SavedFilename = "SuperFile_File1"
		build.Tasks[1].Results[0].DownloadPath = "Dir1/SuperFile_File1"
		//act
		err := build.PublishDownloadedResults("MyStep", filenames, &files)
		//assert
		assert.Error(t, err)
	})
}

func TestDetermineFailureCause(t *testing.T) {
	build := Build{
		Tasks: []task{
			{
				TaskID:      0,
				ResultState: Successful,
				Logs: []logStruct{
					{
						Msgty:   loginfo,
						Logline: "Build successfully initialized",
					},
					{
						Msgty:   loginfo,
						Logline: "2 Plugins will be executed",
					},
				},
			},
			{
				TaskID:      1,
				ResultState: Successful,
				Logs: []logStruct{
					{
						Msgty:   loginfo,
						Logline: "Plugin 1 did something",
					},
				},
			},
			{
				TaskID:      2,
				ResultState: Successful,
				Logs: []logStruct{
					{
						Msgty:   loginfo,
						Logline: "Plugin 2 did something",
					},
				},
			},
		},
	}

	t.Run("TestSuccess", func(t *testing.T) {
		//act
		cause, err := build.DetermineFailureCause()
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "", cause)
	})
	t.Run("TestErronous", func(t *testing.T) {
		//arrange
		errorMessage := "Error: something went wrong, contact your admin :-P"
		errorBuild := Build{}
		errorBuild.Tasks = append(errorBuild.Tasks, build.Tasks[0], task{}, build.Tasks[2])
		errorBuild.Tasks[1].ResultState = Erroneous
		errorBuild.Tasks[1].Logs = append(errorBuild.Tasks[1].Logs, logStruct{Msgty: logerror, Logline: errorMessage})
		//act
		cause, err := errorBuild.DetermineFailureCause()
		//assert
		assert.NoError(t, err)
		assert.Contains(t, cause, errorMessage)
	})
	t.Run("TestAborting", func(t *testing.T) {
		//arrange
		abortMessage := "Aborting: something went wrong, contact your admin :-P"
		abortBuild := Build{}
		abortBuild.Tasks = append(abortBuild.Tasks, build.Tasks[0], build.Tasks[1], task{})
		abortBuild.Tasks[2].ResultState = Aborted
		abortBuild.Tasks[2].Logs = append(abortBuild.Tasks[1].Logs, logStruct{Msgty: logerror, Logline: abortMessage})
		//act
		cause, err := abortBuild.DetermineFailureCause()
		//assert
		assert.NoError(t, err)
		assert.Contains(t, cause, abortMessage)
	})
}
