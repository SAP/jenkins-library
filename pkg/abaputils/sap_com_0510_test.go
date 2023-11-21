package abaputils

import (
	"testing"
	"time"

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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

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
				`{ }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errClone := api.Clone()
		assert.ErrorContains(t, errClone, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCheckoput := api.CheckoutBranch()
		assert.ErrorContains(t, errCheckoput, "Request to ABAP System not successful")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

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

		apiManager := &SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Microsecond}

		api, err := apiManager.GetAPI(con, repo)
		assert.NoError(t, err)
		assert.IsType(t, &SAP_COM_0510{}, api.(*SAP_COM_0510), "API has wrong type")

		errCreateTag := api.CreateTag(Tag{})
		assert.ErrorContains(t, errCreateTag, "No Tag provided")
		assert.Empty(t, api.getUUID(), "API does not cotain correct UUID")
	})
}
