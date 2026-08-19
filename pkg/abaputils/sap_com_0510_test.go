//go:build unit
// +build unit

package abaputils

import (
	"testing"
	"time"

	"errors"

	"github.com/stretchr/testify/assert"
)

var con ConnectionDetailsHTTP
var repo Repository

func init() {

	con.User = "CC_USER"
	con.Password = "123abc"
	con.URL = "https://example.com"

	repo.Name = "/DMO/REPO"
	repo.Branch = "main"

}

func TestRetry(t *testing.T) {
	t.Run("Test retry success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Software component lifecycle activities in progress. Try again later..."} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				nil,
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errAction := api.(*SAP_COM_0510).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.NoError(t, errAction)
		assert.Equal(t, "GUID", api.getUUID(), "API does not cotain correct UUID")

	})

	t.Run("Test retry not allowed", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{"error" : { "code" : "A4C_A2G/224", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				nil,
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errAction := api.(*SAP_COM_0510).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.ErrorContains(t, errAction, "HTTP 400: A4C_A2G/224 - Error Text")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")

	})

	t.Run("Test retry maxSleepTime", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		api.setSleepTimeConfig(time.Nanosecond, 20*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		api.(*SAP_COM_0510).maxRetries = 20

		errAction := api.(*SAP_COM_0510).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.ErrorContains(t, errAction, "HTTP 400: A4C_A2G/228 - Error Text")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")

		assert.Equal(t, 6, len(client.BodyList), "Expected maxSleepTime to limit requests")
	})

	t.Run("Test retry maxRetries", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Error Text"} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		api.setSleepTimeConfig(time.Nanosecond, 999*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		api.(*SAP_COM_0510).maxRetries = 3

		errAction := api.(*SAP_COM_0510).triggerRequest(ConnectionDetailsHTTP{User: "CC_USER", Password: "abc123", URL: "https://example.com/path"}, []byte("{}"))
		assert.ErrorContains(t, errAction, "HTTP 400: A4C_A2G/228 - Error Text")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")

		assert.Equal(t, 5, len(client.BodyList), "Expected maxRetries to limit requests")
	})

}
func TestClone(t *testing.T) {
	t.Run("Test Clone Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errClone := api.Clone()
		assert.NoError(t, errClone)
		assert.Equal(t, "GUID", api.getUUID(), "API does not cotain correct UUID")
	})

	t.Run("Test Clone Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errClone := api.Clone()
		assert.ErrorContains(t, errClone, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
	})

	t.Run("Test Clone Retry", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{"error" : { "code" : "A4C_A2G/228", "message" : { "lang" : "de", "value" : "Software component lifecycle activities in progress. Try again later..."} } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
			ErrorList: []error{
				nil,
				errors.New("HTTP 400"),
				nil,
			},
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		api.setSleepTimeConfig(time.Nanosecond, 120*time.Nanosecond)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errClone := api.Clone()
		assert.NoError(t, errClone)
		assert.Equal(t, "GUID", api.getUUID(), "API does not cotain correct UUID")
	})
}

func TestPull(t *testing.T) {
	t.Run("Test Pull Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errPull := api.Pull()
		assert.NoError(t, errPull)
		assert.Equal(t, "GUID", api.getUUID(), "API does not cotain correct UUID")
	})

	t.Run("Test Pull Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errPull := api.Pull()
		assert.ErrorContains(t, errPull, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
	})
}

func TestCheckout(t *testing.T) {
	t.Run("Test Checkout Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCheckout := api.CheckoutBranch()
		assert.NoError(t, errCheckout)
		assert.Equal(t, "GUID", api.getUUID(), "API does not cotain correct UUID")
	})

	t.Run("Test Checkout Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCheckoput := api.CheckoutBranch()
		assert.ErrorContains(t, errCheckoput, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
	})
}

func TestGetRepo(t *testing.T) {
	t.Run("Test GetRepo Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "sc_name" : "testRepo1", "avail_on_inst" : true, "active_branch": "testBranch1" } }`,
				`{"d" : [] }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		cloned, activeBranch, errAction, _ := api.GetRepository()
		assert.True(t, cloned)
		assert.Equal(t, "testBranch1", activeBranch)
		assert.NoError(t, errAction)
	})
}

func TestCreateTag(t *testing.T) {
	t.Run("Test Tag Success", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{"d" : { "status" : "R", "UUID" : "GUID" } }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{TagName: "myTag", TagDescription: "descr"})
		assert.NoError(t, errCreateTag)
		assert.Equal(t, "GUID", api.getUUID(), "API does not cotain correct UUID")
	})

	t.Run("Test Tag Failure", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{TagName: "myTag", TagDescription: "descr"})
		assert.ErrorContains(t, errCreateTag, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
	})

	t.Run("Test Tag Empty", func(t *testing.T) {

		client := &ClientMock{
			BodyList: []string{
				`{ "d" : {} }`,
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond, Force0510: true}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{})
		assert.ErrorContains(t, errCreateTag, "No Tag provided")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
	})
}

func TestSleepTime(t *testing.T) {
	t.Run("Test Sleep Time", func(t *testing.T) {

		api := SAP_COM_0510{
			retryMaxSleepTime:  120 * time.Nanosecond,
			retryBaseSleepUnit: 1 * time.Nanosecond,
		}

		expectedResults := make([]time.Duration, 12)
		expectedResults[0] = 0
		expectedResults[1] = 1
		expectedResults[2] = 1
		expectedResults[3] = 2
		expectedResults[4] = 3
		expectedResults[5] = 5
		expectedResults[6] = 8
		expectedResults[7] = 13
		expectedResults[8] = 21
		expectedResults[9] = 34
		expectedResults[10] = 55
		expectedResults[11] = 89
		results := make([]time.Duration, 12)
		var err error

		for i := 0; i <= 11; i++ {

			results[i], err = api.getSleepTime(i)
			assert.NoError(t, err)
		}
		assert.ElementsMatch(t, expectedResults, results)

		_, err = api.getSleepTime(-10)
		assert.Error(t, err)

		_, err = api.getSleepTime(12)
		assert.ErrorContains(t, err, "Exceeded max sleep time")
	})
}

func TestTimeConverter(t *testing.T) {

	api := SAP_COM_0510{}

	t.Run("Test example time", func(t *testing.T) {
		inputDate := "/Date(1585576809000+0000)/"
		expectedDate := "2020-03-30 14:00:09 +0000 UTC"
		result := api.ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test Unix time", func(t *testing.T) {
		inputDate := "/Date(0000000000000+0000)/"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := api.ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
	t.Run("Test unexpected format", func(t *testing.T) {
		inputDate := "/Date(0012300000001+0000)/"
		expectedDate := "1970-01-01 00:00:00 +0000 UTC"
		result := api.ConvertTime(inputDate)
		assert.Equal(t, expectedDate, result.String(), "Dates do not match after conversion")
	})
}
