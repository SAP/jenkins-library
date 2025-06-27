//go:build unit

package cmd

import (
	"testing"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestPolling(t *testing.T) {
	t.Run("Run polling", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&abapbuild.ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		timeout := time.Duration(600 * time.Second)
		pollInterval := time.Duration(1 * time.Second)
		err := polling(buildsWithRepo, timeout, pollInterval)
		assert.NoError(t, err)
		assert.Equal(t, abapbuild.Finished, buildsWithRepo[0].build.RunState)
	})
}
func TestStartingConfirm(t *testing.T) {
	t.Run("Run starting", func(t *testing.T) {
		client := &abapbuild.ClMock{
			Token: "MyToken",
		}
		conn := new(abapbuild.Connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:        "RepoA",
			Version:     "0001",
			PackageName: "Package",
			PackageType: "AOI",
			SpLevel:     "0000",
			PatchLevel:  "0000",
			Status:      "P",
			Namespace:   "/DEMO/",
		}
		repos = append(repos, repo)
		repo.Status = "R"
		repo.InBuildScope = true
		repos = append(repos, repo)

		builds, err := startingConfirm(repos, *conn, time.Duration(0*time.Second))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(builds))
		assert.Equal(t, abapbuild.Accepted, builds[0].build.RunState)
	})
}

func TestStartingConfirmInvalidInput(t *testing.T) {
	t.Run("Run starting", func(t *testing.T) {
		client := &abapbuild.ClMock{
			Token: "MyToken",
		}
		conn := new(abapbuild.Connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:         "RepoA",
			InBuildScope: true,
		}
		repos = append(repos, repo)
		_, err := startingConfirm(repos, *conn, time.Duration(0*time.Second))
		assert.Error(t, err)
	})
}

func TestStartingConfirmReleaseFailed(t *testing.T) {
	t.Run("Run starting release failed", func(t *testing.T) {
		client := &abapbuild.ClMock{
			Token: "MyToken",
		}
		conn := new(abapbuild.Connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:         "RepoA",
			Version:      "0001",
			PackageName:  "Package",
			PackageType:  "AOI",
			SpLevel:      "0000",
			PatchLevel:   "0000",
			Status:       "L",
			Namespace:    "/DEMO/",
			InBuildScope: true,
		}
		repos = append(repos, repo)
		repo.Status = "R"
		repos = append(repos, repo)

		builds, err := startingConfirm(repos, *conn, time.Duration(0*time.Second))
		assert.Error(t, err)
		assert.Equal(t, 1, len(builds))
		assert.Equal(t, abapbuild.Accepted, builds[0].build.RunState)
	})
}
