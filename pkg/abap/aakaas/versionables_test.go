//go:build unit

package aakaas

import (
	"testing"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/stretchr/testify/assert"
)

func TestCvResolve(t *testing.T) {
	//arrange
	conn := new(abapbuild.Connector)
	mc := abapbuild.NewMockClient()
	conn.Client = &mc
	vers := versionable{}

	t.Run("Factory Success", func(t *testing.T) {
		//act
		err := vers.constructVersionable("DummyComp", "1.2.3", *conn, "")
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "DummyComp", vers.Name)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0002", vers.TechSpLevel)
		assert.Equal(t, "0003", vers.TechPatchLevel)
		assert.Equal(t, "1.2.3", vers.Version)
	})
	t.Run("Factory No Name", func(t *testing.T) {
		err := vers.constructVersionable("", "1.2.3", *conn, "")
		assert.Error(t, err)
	})
	t.Run("Factory Version too long", func(t *testing.T) {
		err := vers.constructVersionable("DummyComp", "1.0.0.0", *conn, "")
		assert.Error(t, err)
	})
	t.Run("Factory Version too short", func(t *testing.T) {

		err := vers.constructVersionable("DummyComp", "1.0", *conn, "")
		assert.Error(t, err)
	})
	t.Run("ComponentVersion NEXT Release Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetReleaseExisting)
		err := vers.constructVersionable("DummyComp", wildCardNext+".0.0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "2", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "2.0.0", vers.Version)
	})
	t.Run("ComponentVersion NEXT Release Non Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetReleaseNonExisting)
		err := vers.constructVersionable("DummyComp", wildCardNext+".0.0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.0.0", vers.Version)
	})
	t.Run("ComponentVersion NEXT SP Level Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetSpLevelExisting)
		err := vers.constructVersionable("DummyComp", "1."+wildCardNext+".0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0008", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.8.0", vers.Version)
	})
	t.Run("ComponentVersion NEXT SP Level Non Existing", func(t *testing.T) {
		//This one should lead to an error later on as AOI is needed - anyway we can't just produce a differen package then customized...
		mc.AddData(testDataAakaasCVGetSpLevelNonExisting)
		err := vers.constructVersionable("DummyComp", "1."+wildCardNext+".0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0001", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.1.0", vers.Version)
	})
	t.Run("ComponentVersion NEXT Patch Level Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetPatchLevelExisting)
		err := vers.constructVersionable("DummyComp", "1.3."+wildCardNext, *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0003", vers.TechSpLevel)
		assert.Equal(t, "0047", vers.TechPatchLevel)
		assert.Equal(t, "1.3.47", vers.Version)
	})
	t.Run("ComponentVersion NEXT Patch Level Non Existing", func(t *testing.T) {
		//This one should lead to an error later on as AOI is needed - anyway we can't just produce a differen package then customized...
		mc.AddData(testDataAakaasCVGetPatchLevelNonExisting)
		err := vers.constructVersionable("DummyComp", "1.3."+wildCardNext, *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0003", vers.TechSpLevel)
		assert.Equal(t, "0001", vers.TechPatchLevel)
		assert.Equal(t, "1.3.1", vers.Version)
	})

	t.Run("Product Version NEXT Release Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasPVGetReleaseExisting)
		err := vers.constructVersionable("DummyProd", wildCardNext+".0.0", *conn, pvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterPV)
		assert.NoError(t, err)
		assert.Equal(t, "2", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "2.0.0", vers.Version)
	})

	t.Run("Product Version NEXT Release Non Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasPVGetReleaseNonExisting)
		err := vers.constructVersionable("DummyProd", wildCardNext+".0.0", *conn, pvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterPV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.0.0", vers.Version)
	})

	t.Run("Component Version MAX Release existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetReleaseExisting)
		err := vers.constructVersionable("DummyComp", wildCardMax+".0.0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.0.0", vers.Version)
	})

	t.Run("Component Version MAX Release non existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetReleaseNonExisting)
		err := vers.constructVersionable("DummyComp", wildCardMax+".0.0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "0", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "0.0.0", vers.Version)
	})

	t.Run("Component Version MAX SP Level existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetSpLevelExisting)
		err := vers.constructVersionable("DummyComp", "1."+wildCardMax+".0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0007", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.7.0", vers.Version)
	})

	t.Run("Component Version MAX SP Level non existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetSpLevelNonExisting)
		err := vers.constructVersionable("DummyComp", "1."+wildCardMax+".0", *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.0.0", vers.Version)
	})

	t.Run("Component Version MAX Patch Level existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetPatchLevelExisting)
		err := vers.constructVersionable("DummyComp", "1.3."+wildCardMax, *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0003", vers.TechSpLevel)
		assert.Equal(t, "0046", vers.TechPatchLevel)
		assert.Equal(t, "1.3.46", vers.Version)
	})

	t.Run("Component Version MAX Patch Level non existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetPatchLevelNonExisting)
		err := vers.constructVersionable("DummyComp", "1.3."+wildCardMax, *conn, cvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterCV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0003", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.3.0", vers.Version)
	})

	t.Run("Product Version MAX Release existing", func(t *testing.T) {
		mc.AddData(testDataAakaasPVGetReleaseExisting)
		err := vers.constructVersionable("DummyProd", wildCardMax+".0.0", *conn, pvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterPV)
		assert.NoError(t, err)
		assert.Equal(t, "1", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "1.0.0", vers.Version)
	})

	t.Run("Product Version MAX Release non existing", func(t *testing.T) {
		mc.AddData(testDataAakaasPVGetReleaseNonExisting)
		err := vers.constructVersionable("DummyProd", wildCardMax+".0.0", *conn, pvQueryURL)
		assert.NoError(t, err)
		err = vers.resolveWildCards(statusFilterPV)
		assert.NoError(t, err)
		assert.Equal(t, "0", vers.TechRelease)
		assert.Equal(t, "0000", vers.TechSpLevel)
		assert.Equal(t, "0000", vers.TechPatchLevel)
		assert.Equal(t, "0.0.0", vers.Version)
	})
}
