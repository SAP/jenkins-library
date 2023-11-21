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
