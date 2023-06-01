//go:build unit
// +build unit

package aakaas

import (
	"testing"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestInitPV(t *testing.T) {
	conn := new(abapbuild.Connector)
	conn.Client = &abaputils.ClientMock{}
	prodVers := abaputils.AddonDescriptor{
		AddonProduct:     "/DRNMSPC/PRD01",
		AddonVersionYAML: "3.2.1",
	}
	var pv ProductVersion

	t.Run("test init", func(t *testing.T) {
		pv.ConstructProductversion(prodVers, *conn)
		assert.Equal(t, "/DRNMSPC/PRD01", pv.Name)
		assert.Equal(t, "3.2.1", pv.Version)
	})

	t.Run("test validate - success", func(t *testing.T) {
		conn.Client = &abaputils.ClientMock{
			Body: ResponseCheckPV,
		}
		pv.ConstructProductversion(prodVers, *conn)
		err := pv.ValidateAndResolveVersionFields()
		assert.NoError(t, err)
		assert.Equal(t, "0003", pv.TechRelease)
		assert.Equal(t, "0002", pv.TechSpLevel)
		assert.Equal(t, "0001", pv.TechPatchLevel)
	})

	t.Run("test validate - with error", func(t *testing.T) {
		conn.Client = &abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Validation failed"),
		}
		pv.ConstructProductversion(prodVers, *conn)
		err := pv.ValidateAndResolveVersionFields()
		assert.Error(t, err)
	})

	t.Run("test copyFieldsToRepo", func(t *testing.T) {
		pv.TechRelease = "0003"
		pv.TechSpLevel = "0002"
		pv.TechPatchLevel = "0001"
		pv.CopyVersionFieldsToDescriptor(&prodVers)
		assert.Equal(t, "0003", prodVers.AddonVersion)
		assert.Equal(t, "0002", prodVers.AddonSpsLevel)
		assert.Equal(t, "0001", prodVers.AddonPatchLevel)
	})
}
