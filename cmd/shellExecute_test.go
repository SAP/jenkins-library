package cmd

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SAP/jenkins-library/pkg/mock"
)

type shellExecuteMockUtils struct {
	config *shellExecuteOptions
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
	downloadError error
	filename      string
	header        http.Header
	url           string
}

func (f *shellExecuteMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	if f.downloadError != nil {
		return f.downloadError
	}
	f.url = url
	f.filename = filename
	f.header = header
	return nil
}

func newShellExecuteTestsUtils() *shellExecuteMockUtils {
	utils := shellExecuteMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utils
}

func (v *shellExecuteMockUtils) GetConfig() *shellExecuteOptions {
	return v.config
}

func TestRunShellExecute(t *testing.T) {

	t.Run("negative case - script isn't present", func(t *testing.T) {
		c := &shellExecuteOptions{
			Sources: []string{"path/to/script.sh"},
		}
		u := newShellExecuteTestsUtils()

		err := runShellExecute(c, nil, u)
		assert.EqualError(t, err, "the script 'path/to/script.sh' could not be found")
	})

	t.Run("success case - script run successfully", func(t *testing.T) {
		o := &shellExecuteOptions{
			Sources: []string{"path/script.sh"},
		}

		u := newShellExecuteTestsUtils()
		u.AddFile("path/script.sh", []byte(`echo dummy`))

		err := runShellExecute(o, nil, u)
		assert.Equal(t, "path/script.sh", u.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{}, u.ExecMockRunner.Calls[0].Params)
		assert.NoError(t, err)
	})

	t.Run("success case - download script header gets added", func(t *testing.T) {
		o := &shellExecuteOptions{
			Sources:     []string{"https://myScriptLocation/myScript.sh"},
			GithubToken: "dummy@12345",
		}
		u := newShellExecuteTestsUtils()

		runShellExecute(o, nil, u)

		assert.Equal(t, http.Header{"Accept": []string{"application/vnd.github.v3.raw"}, "Authorization": []string{"Token dummy@12345"}}, u.header)
	})

	t.Run("success case - single positional script arguments gets added to the correct script", func(t *testing.T) {
		o := &shellExecuteOptions{
			Sources:         []string{"path1/script1.sh", "path2/script2.sh"},
			ScriptArguments: []string{"arg1", "arg2"},
		}

		u := newShellExecuteTestsUtils()
		u.AddFile("path1/script1.sh", []byte(`echo dummy1`))
		u.AddFile("path2/script2.sh", []byte(`echo dummy2`))

		err := runShellExecute(o, nil, u)

		assert.Equal(t, "path1/script1.sh", u.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"arg1"}, u.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "path2/script2.sh", u.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"arg2"}, u.ExecMockRunner.Calls[1].Params)
		assert.NoError(t, err)
	})

	t.Run("success case - multiple positional script arguments gets added to the correct script", func(t *testing.T) {
		o := &shellExecuteOptions{
			Sources:         []string{"path1/script1.sh", "path2/script2.sh"},
			ScriptArguments: []string{"arg1,arg2", "arg3,arg4"},
		}

		u := newShellExecuteTestsUtils()
		u.AddFile("path1/script1.sh", []byte(`echo dummy1`))
		u.AddFile("path2/script2.sh", []byte(`echo dummy2`))

		err := runShellExecute(o, nil, u)

		assert.Equal(t, "path1/script1.sh", u.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{"arg1", "arg2"}, u.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "path2/script2.sh", u.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"arg3", "arg4"}, u.ExecMockRunner.Calls[1].Params)
		assert.NoError(t, err)
	})

	t.Run("success case - no argument for script 1, single argument for script 2 and multiple argument for script3 gets added to the correct script", func(t *testing.T) {
		o := &shellExecuteOptions{
			Sources:         []string{"path1/script1.sh", "path2/script2.sh", "path3/script3.sh"},
			ScriptArguments: []string{"", "arg2.1", "arg3.1,arg3.2"},
		}

		u := newShellExecuteTestsUtils()
		u.AddFile("path1/script1.sh", []byte(`echo dummy1`))
		u.AddFile("path2/script2.sh", []byte(`echo dummy2`))
		u.AddFile("path3/script3.sh", []byte(`echo dummy3`))

		err := runShellExecute(o, nil, u)

		assert.Equal(t, "path1/script1.sh", u.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, []string{}, u.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, "path2/script2.sh", u.ExecMockRunner.Calls[1].Exec)
		assert.Equal(t, []string{"arg2.1"}, u.ExecMockRunner.Calls[1].Params)
		assert.Equal(t, "path3/script3.sh", u.ExecMockRunner.Calls[2].Exec)
		assert.Equal(t, []string{"arg3.1", "arg3.2"}, u.ExecMockRunner.Calls[2].Params)
		assert.NoError(t, err)
	})
}
