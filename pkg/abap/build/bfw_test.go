package build

import (
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
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
