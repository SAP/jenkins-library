package aakaas

import (
	"testing"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestInitCV(t *testing.T) {
	conn := new(abapbuild.Connector)
	conn.Client = &abaputils.ClientMock{}
	repo := abaputils.Repository{
		Name:        "/DRNMSPC/COMP01",
		VersionYAML: "1.2.3",
	}
	var c ComponentVersion

	t.Run("test init", func(t *testing.T) {
		c.ConstructComponentVersion(repo, *conn)
		assert.Equal(t, "/DRNMSPC/COMP01", c.Name)
		assert.Equal(t, "1.2.3", c.Version)
	})

	t.Run("test validate - success", func(t *testing.T) {
		conn.Client = &abaputils.ClientMock{
			Body: ResponseCheckCVs,
		}
		c.ConstructComponentVersion(repo, *conn)

		err := c.Validate()

		assert.NoError(t, err)
		assert.Equal(t, "0001", c.TechRelease)
		assert.Equal(t, "0002", c.TechSpLevel)
		assert.Equal(t, "0003", c.TechPatchLevel)
	})

	t.Run("test validate - with error", func(t *testing.T) {
		conn.Client = &abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Validation failed"),
		}
		c.ConstructComponentVersion(repo, *conn)

		err := c.Validate()

		assert.Error(t, err)
	})

	t.Run("test copyFieldsToRepo", func(t *testing.T) {

		var c ComponentVersion
		c.TechRelease = "0001"
		c.TechSpLevel = "0002"
		c.TechPatchLevel = "0003"
		c.CopyVersionFieldsToRepo(&repo)
		assert.Equal(t, "0001", repo.Version)
		assert.Equal(t, "0002", repo.SpLevel)
		assert.Equal(t, "0003", repo.PatchLevel)
	})
}
