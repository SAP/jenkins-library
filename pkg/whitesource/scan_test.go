package whitesource

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppendScannedProjectVersion(t *testing.T) {
	t.Parallel()
	t.Run("single module", func(t *testing.T) {
		// init
		scan := NewScan(ScanOptions{ProductVersion: "1"})
		// test
		err := scan.AppendScannedProjectVersion("module-a - 1")
		// assert
		assert.NoError(t, err)
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
		_, exists := scan.scanTimes["module-a - 1"]
		assert.True(t, exists)
	})
	t.Run("two modules", func(t *testing.T) {
		// init
		scan := NewScan(ScanOptions{ProductVersion: "1"})
		// test
		err1 := scan.AppendScannedProjectVersion("module-a - 1")
		err2 := scan.AppendScannedProjectVersion("module-b - 1")
		// assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		expected["module-b - 1"] = Project{Name: "module-b - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
		_, exists := scan.scanTimes["module-b - 1"]
		assert.True(t, exists)
	})
	t.Run("module without version", func(t *testing.T) {
		// init
		scan := NewScan(ScanOptions{ProductVersion: "1"})
		// test
		err := scan.AppendScannedProjectVersion("module-a")
		// assert
		assert.EqualError(t, err, "projectName is expected to include the product version")
		assert.Len(t, scan.scannedProjects, 0)
	})
	t.Run("duplicate module", func(t *testing.T) {
		// init
		scan := NewScan(ScanOptions{ProductVersion: "1"})
		// test
		err1 := scan.AppendScannedProjectVersion("module-a - 1")
		err2 := scan.AppendScannedProjectVersion("module-a - 1")
		// assert
		assert.NoError(t, err1)
		assert.EqualError(t, err2, "project with name 'module-a - 1' was already scanned")
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
		assert.Len(t, scan.scanTimes, 1)
	})
}

func TestAppendScannedProject(t *testing.T) {
	t.Parallel()
	t.Run("product version is appended", func(t *testing.T) {
		// init
		scan := NewScan(ScanOptions{ProductVersion: "1"})
		// test
		err := scan.AppendScannedProject("module-a")
		// assert
		assert.NoError(t, err)
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
	})
}

func TestNewScan(t *testing.T) {
	t.Parallel()
	t.Run("options are transferred", func(t *testing.T) {
		// init
		options := ScanOptions{
			AggregateProjectName: "project",
			ProductVersion:       "1",
		}
		// test
		scan := NewScan(options)
		// assert
		assert.Equal(t, &Scan{aggregateProjectName: "project", productVersion: "1"}, scan)
	})
}
