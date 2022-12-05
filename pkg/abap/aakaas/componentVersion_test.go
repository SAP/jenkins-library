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

	t.Run("ComponentVersion Factory Success", func(t *testing.T) {
		//act
		cv, err := NewComponentVersion("DummyComp", "1.2.3", *conn)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "DummyComp", cv.Name)
		assert.Equal(t, "1", cv.Release)
		assert.Equal(t, "0002", cv.SpLevel)
		assert.Equal(t, "0003", cv.PatchLevel)
	})
	t.Run("ComponentVersion Factory No Name", func(t *testing.T) {
		_, err := NewComponentVersion("", "1.0.0", *conn)
		assert.Error(t, err)
	})
	t.Run("ComponentVersion Factory Version too long", func(t *testing.T) {
		_, err := NewComponentVersion("DummyComp", "1.0.0.0", *conn)
		assert.Error(t, err)
	})
	t.Run("ComponentVersion Factory Version too short", func(t *testing.T) {
		_, err := NewComponentVersion("DummyComp", "1.0", *conn)
		assert.Error(t, err)
	})
	t.Run("ComponentVersion NEXT Release Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetReleaseExisting)
		cv, err := NewComponentVersion("DummyComp", wildCard+".0.0", *conn)
		assert.NoError(t, err)
		err = cv.ResolveNext()
		assert.NoError(t, err)
		assert.Equal(t, "2", cv.Release)
		assert.Equal(t, "0000", cv.SpLevel)
		assert.Equal(t, "0000", cv.PatchLevel)
	})
	t.Run("ComponentVersion NEXT Release Non Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetReleaseNonExisting)
		cv, err := NewComponentVersion("DummyComp", wildCard+".0.0", *conn)
		assert.NoError(t, err)
		err = cv.ResolveNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", cv.Release)
		assert.Equal(t, "0000", cv.SpLevel)
		assert.Equal(t, "0000", cv.PatchLevel)
	})
	t.Run("ComponentVersion NEXT SP Level Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetSpLevelExisting)
		cv, err := NewComponentVersion("DummyComp", "1."+wildCard+".0", *conn)
		assert.NoError(t, err)
		err = cv.ResolveNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", cv.Release)
		assert.Equal(t, "0008", cv.SpLevel)
		assert.Equal(t, "0000", cv.PatchLevel)
	})
	t.Run("ComponentVersion NEXT SP Level Non Existing", func(t *testing.T) {
		//This one should lead to an error later on as AOI is needed - anyway we can't just produce a differen package then customized...
		mc.AddData(testDataAakaasCVGetSpLevelNonExisting)
		cv, err := NewComponentVersion("DummyComp", "1."+wildCard+".0", *conn)
		assert.NoError(t, err)
		err = cv.ResolveNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", cv.Release)
		assert.Equal(t, "0001", cv.SpLevel)
		assert.Equal(t, "0000", cv.PatchLevel)
	})
	t.Run("ComponentVersion NEXT Patch Level Existing", func(t *testing.T) {
		mc.AddData(testDataAakaasCVGetPatchLevelExisting)
		cv, err := NewComponentVersion("DummyComp", "1.3."+wildCard, *conn)
		assert.NoError(t, err)
		err = cv.ResolveNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", cv.Release)
		assert.Equal(t, "0003", cv.SpLevel)
		assert.Equal(t, "0047", cv.PatchLevel)
	})
	t.Run("ComponentVersion NEXT Patch Level Non Existing", func(t *testing.T) {
		//This one should lead to an error later on as AOI is needed - anyway we can't just produce a differen package then customized...
		mc.AddData(testDataAakaasCVGetPatchLevelNonExisting)
		cv, err := NewComponentVersion("DummyComp", "1.3."+wildCard, *conn)
		assert.NoError(t, err)
		err = cv.ResolveNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", cv.Release)
		assert.Equal(t, "0003", cv.SpLevel)
		assert.Equal(t, "0001", cv.PatchLevel)
	})
}

var testDataAakaasCVGetReleaseExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechSpLevel+eq+%270000%27+and+TechPatchLevel+eq+%270000%27&%24format=json&%24orderby=TechRelease+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel&%24top=1`,
	Body: `{
		"d": {
			"results": [
				{
					"Name": "DummyComp",
					"Version": "1.0.0",
					"TechRelease": "1",
					"TechSpLevel": "0000",
					"TechPatchLevel": "0000"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetReleaseNonExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechSpLevel+eq+%270000%27+and+TechPatchLevel+eq+%270000%27&%24format=json&%24orderby=TechRelease+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel&%24top=1`,
	Body: `{
		"d": {
			"results": []
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetSpLevelExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechPatchLevel+eq+%270000%27&%24format=json&%24orderby=TechSpLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel&%24top=1`,
	Body: `{
		"d": {
			"results": [
				{
					"Name": "DummyComp",
					"Version": "1.7.0",
					"TechRelease": "1",
					"TechSpLevel": "0007",
					"TechPatchLevel": "0000"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetSpLevelNonExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechPatchLevel+eq+%270000%27&%24format=json&%24orderby=TechSpLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel&%24top=1`,
	Body: `{
		"d": {
			"results": []
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetPatchLevelExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechSpLevel+eq+%270003%27&%24format=json&%24orderby=TechPatchLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel&%24top=1`,
	Body: `{
		"d": {
			"results": [
				{
					"Name": "DummyComp",
					"Version": "1.3.46",
					"TechRelease": "1",
					"TechSpLevel": "0003",
					"TechPatchLevel": "0046"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetPatchLevelNonExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechSpLevel+eq+%270003%27&%24format=json&%24orderby=TechPatchLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel&%24top=1`,
	Body: `{
		"d": {
			"results": []
		}
	}`,
	StatusCode: 200,
}
