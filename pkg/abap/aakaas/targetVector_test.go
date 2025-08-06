package aakaas

import (
	"testing"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestTargetVectorInitExisting(t *testing.T) {
	t.Run("ID is set", func(t *testing.T) {
		//arrange
		id := "dummyID"
		targetVector := new(TargetVector)
		//act
		targetVector.InitExisting(id)
		//assert
		assert.Equal(t, id, targetVector.ID)
	})
}

func TestTargetVectorInitNew(t *testing.T) {
	t.Run("Ensure values not initial", func(t *testing.T) {
		//arrange
		addonDescriptor := abaputils.AddonDescriptor{
			AddonProduct:    "dummy",
			AddonVersion:    "dummy",
			AddonSpsLevel:   "dummy",
			AddonPatchLevel: "dummy",
			TargetVectorID:  "dummy",
			Repositories: []abaputils.Repository{
				{
					Name:        "dummy",
					Version:     "dummy",
					SpLevel:     "dummy",
					PatchLevel:  "dummy",
					PackageName: "dummy",
				},
			},
		}
		targetVector := new(TargetVector)
		//act
		err := targetVector.InitNew(&addonDescriptor)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "dummy", targetVector.ProductVersion)
	})
	t.Run("Fail if values initial", func(t *testing.T) {
		//arrange
		addonDescriptor := abaputils.AddonDescriptor{}
		targetVector := new(TargetVector)
		//act
		err := targetVector.InitNew(&addonDescriptor)
		//assert
		assert.Error(t, err)
	})
}

func TestTargetVectorGet(t *testing.T) {
	//arrange global
	targetVector := new(TargetVector)
	conn := new(abapbuild.Connector)

	t.Run("Ensure error if ID is initial", func(t *testing.T) {
		//arrange
		targetVector.ID = ""
		//act
		err := targetVector.GetTargetVector(conn)
		//assert
		assert.Error(t, err)
	})
	t.Run("Normal Get Test Success", func(t *testing.T) {
		//arrange
		targetVector.ID = "W7Q00207512600000353"
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSGetTVPublishTestSuccess)
		conn.Client = &mc
		//act
		err := targetVector.GetTargetVector(conn)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, TargetVectorPublishStatusSuccess, TargetVectorStatus(targetVector.PublishStatus))
		assert.Equal(t, TargetVectorStatusTest, TargetVectorStatus(targetVector.Status))
	})
	t.Run("Error Get", func(t *testing.T) {
		//arrange
		targetVector.ID = "W7Q00207512600000353"
		mc := abapbuild.NewMockClient()
		conn.Client = &mc
		//act
		err := targetVector.GetTargetVector(conn)
		//assert
		assert.Error(t, err)
	})
}

func TestTargetVectorPollForStatus(t *testing.T) {
	//arrange global
	targetVector := new(TargetVector)
	conn := new(abapbuild.Connector)
	conn.MaxRuntime = time.Duration(1 * time.Second)
	conn.PollingInterval = time.Duration(1 * time.Microsecond)

	t.Run("Normal Poll", func(t *testing.T) {
		//arrange
		targetVector.ID = "W7Q00207512600000353"
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSGetTVPublishRunning)
		mc.AddData(AAKaaSGetTVPublishTestSuccess)
		conn.Client = &mc
		//act
		err := targetVector.PollForStatus(conn, TargetVectorStatusTest)
		//assert
		assert.NoError(t, err)
	})
}

func TestTargetVectorCreate(t *testing.T) {
	//arrange global
	targetVector := new(TargetVector)
	conn := new(abapbuild.Connector)

	addonDescriptor := abaputils.AddonDescriptor{
		AddonProduct:    "dummy",
		AddonVersion:    "dummy",
		AddonSpsLevel:   "dummy",
		AddonPatchLevel: "dummy",
		TargetVectorID:  "dummy",
		Repositories: []abaputils.Repository{
			{
				Name:        "dummy",
				Version:     "dummy",
				SpLevel:     "dummy",
				PatchLevel:  "dummy",
				PackageName: "dummy",
			},
		},
	}

	t.Run("Create Success", func(t *testing.T) {
		//arrange
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSHead)
		mc.AddData(AAKaaSTVCreatePost)
		errInitConn := conn.InitAAKaaS("", "dummyUser", "dummyPassword", &mc, "", "", "")
		assert.NoError(t, errInitConn)

		errInitTV := targetVector.InitNew(&addonDescriptor)
		assert.NoError(t, errInitTV)
		//act
		err := targetVector.CreateTargetVector(conn)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "W7Q00207512600000262", targetVector.ID)
	})
}

func TestTargetVectorPublish(t *testing.T) {
	//arrange global
	targetVector := new(TargetVector)
	conn := new(abapbuild.Connector)

	t.Run("Publish Test", func(t *testing.T) {
		//arrange
		targetVector.ID = "W7Q00207512600000353"
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSHead)
		mc.AddData(AAKaaSTVPublishTestPost)
		errInitConn := conn.InitAAKaaS("", "dummyUser", "dummyPassword", &mc, "", "", "")
		assert.NoError(t, errInitConn)

		//act
		err := targetVector.PublishTargetVector(conn, TargetVectorStatusTest)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, string(TargetVectorPublishStatusRunning), targetVector.PublishStatus)
	})
}
