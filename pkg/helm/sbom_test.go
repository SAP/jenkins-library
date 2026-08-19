//go:build unit
// +build unit

package helm

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
)

func TestRunHelmReadChartMetadata(t *testing.T) {
	t.Run("valid Chart.yaml is parsed", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(
			"apiVersion: v2\nname: foo\nversion: 1.2.3\nappVersion: 4.5.6\n"))

		meta, err := readChartMetadata("chart", files)

		require.NoError(t, err)
		assert.Equal(t, "foo", meta.Name)
		assert.Equal(t, "1.2.3", meta.Version)
		assert.Equal(t, "4.5.6", meta.AppVersion)
	})

	t.Run("missing Chart.yaml returns a read error", func(t *testing.T) {
		files := &mock.FilesMock{}

		meta, err := readChartMetadata("chart", files)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read chart/Chart.yaml")
		assert.Nil(t, meta)
	})

	t.Run("malformed Chart.yaml returns a parse error", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte("name: foo\n  bad: : indentation"))

		meta, err := readChartMetadata("chart", files)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse chart/Chart.yaml")
		assert.Nil(t, meta)
	})
}

func TestRunHelmReadChartLock(t *testing.T) {
	t.Run("valid Chart.lock returns its resolved dependencies", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.lock", []byte(
			"dependencies:\n"+
				"  - name: common\n    version: 2.3.1\n    repository: https://charts.example.com\n"+
				"  - name: postgresql\n    version: 12.1.9\n    repository: https://charts.example.com\n"))

		deps := readChartLock("chart", files)

		require.Len(t, deps, 2)
		assert.Equal(t, "common", deps[0].Name)
		assert.Equal(t, "2.3.1", deps[0].Version)
		assert.Equal(t, "postgresql", deps[1].Name)
		assert.Equal(t, "12.1.9", deps[1].Version)
	})

	t.Run("missing Chart.lock returns nil", func(t *testing.T) {
		files := &mock.FilesMock{}

		assert.Nil(t, readChartLock("chart", files))
	})

	t.Run("malformed Chart.lock returns nil", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.lock", []byte("dependencies: : broken"))

		assert.Nil(t, readChartLock("chart", files))
	})

	t.Run("Chart.lock with no dependencies returns an empty list", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.lock", []byte("dependencies: []\n"))

		assert.Empty(t, readChartLock("chart", files))
	})
}

func TestRunHelmNewChartBOM(t *testing.T) {
	t.Run("root component carries the chart identity and helm PURL", func(t *testing.T) {
		meta := &versioning.Metadata{Name: "foo", Version: "1.2.3", AppVersion: "4.5.6"}

		bom := newChartBOM(meta, "chart", nil, &mock.FilesMock{})

		require.NotNil(t, bom.Metadata)
		require.NotNil(t, bom.Metadata.Component)
		assert.Equal(t, cdx.ComponentTypeApplication, bom.Metadata.Component.Type)
		assert.Equal(t, "foo", bom.Metadata.Component.Name)
		assert.Equal(t, "1.2.3", bom.Metadata.Component.Version)
		assert.Equal(t, "pkg:helm/foo@1.2.3", bom.Metadata.Component.PackageURL)
		assert.Nil(t, bom.Components, "a chart with no deps/images must have no components")
	})

	t.Run("gateway-required fields are populated", func(t *testing.T) {
		meta := &versioning.Metadata{Name: "foo", Version: "1.2.3"}

		bom := newChartBOM(meta, "chart", nil, &mock.FilesMock{})

		// The SBOM gateway requires a serialNumber (urn:uuid), a metadata
		// timestamp, and a bom-ref on the root component.
		assert.Regexp(t, `^urn:uuid:[0-9a-f-]{36}$`, bom.SerialNumber)
		require.NotNil(t, bom.Metadata)
		assert.NotEmpty(t, bom.Metadata.Timestamp, "metadata timestamp is required")
		assert.Equal(t, "pkg:helm/foo@1.2.3", bom.Metadata.Component.BOMRef, "root component needs a bom-ref")
		// The chart SBOM must be pinned to CycloneDX schema version 1.4.
		assert.Equal(t, "1.4", bom.SpecVersion)
		assert.Equal(t, "http://cyclonedx.org/schema/bom/1.4", bom.XMLNS)
	})

	t.Run("dependencies become library components", func(t *testing.T) {
		meta := &versioning.Metadata{
			Name: "foo", Version: "1.2.3",
			Dependencies: []*chart.Dependency{
				{Name: "common", Version: "2.0.0"},
				{Name: "postgresql", Version: "12.1.0"},
			},
		}

		bom := newChartBOM(meta, "chart", nil, &mock.FilesMock{})

		require.NotNil(t, bom.Components)
		require.Len(t, *bom.Components, 2)
		assert.Equal(t, cdx.ComponentTypeLibrary, (*bom.Components)[0].Type)
		assert.Equal(t, "pkg:helm/common@2.0.0", (*bom.Components)[0].PackageURL)
		assert.Equal(t, "pkg:helm/postgresql@12.1.0", (*bom.Components)[1].PackageURL)
	})

	t.Run("Chart.lock resolved versions win over metadata dependencies", func(t *testing.T) {
		meta := &versioning.Metadata{
			Name: "foo", Version: "1.2.3",
			Dependencies: []*chart.Dependency{{Name: "common", Version: "^2.0.0"}},
		}
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.lock", []byte(
			"dependencies:\n  - name: common\n    version: 2.3.1\n"))

		bom := newChartBOM(meta, "chart", nil, files)

		require.NotNil(t, bom.Components)
		require.Len(t, *bom.Components, 1)
		assert.Equal(t, "pkg:helm/common@2.3.1", (*bom.Components)[0].PackageURL)
	})

	t.Run("images become container components, unparseable ones skipped", func(t *testing.T) {
		meta := &versioning.Metadata{Name: "foo", Version: "1.2.3"}
		images := []string{"registry.example.com/app:9.9.9", "bad name"}

		bom := newChartBOM(meta, "chart", images, &mock.FilesMock{})

		require.NotNil(t, bom.Components)
		require.Len(t, *bom.Components, 1, "the unparseable image must be skipped")
		assert.Equal(t, cdx.ComponentTypeContainer, (*bom.Components)[0].Type)
		assert.Equal(t, "pkg:oci/app@9.9.9", (*bom.Components)[0].PackageURL)
	})
}

func TestRunHelmWriteBOM(t *testing.T) {
	t.Run("encodes a conformant BOM as CycloneDX XML and writes it", func(t *testing.T) {
		bom := newChartBOM(&versioning.Metadata{Name: "foo", Version: "1.2.3"}, "chart", nil, &mock.FilesMock{})
		files := &mock.FilesMock{}

		err := writeBOM(bom, "bom-helm.xml", files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		assert.Contains(t, string(content), "pkg:helm/foo@1.2.3")
		assert.Contains(t, string(content), "http://cyclonedx.org/schema/bom/1.4")
	})

	t.Run("non-conformant BOM is rejected before writing", func(t *testing.T) {
		// A bare BOM (no serialNumber/timestamp/xmlns) must fail validation and
		// never reach the filesystem.
		bom := cdx.NewBOM()
		bom.Metadata = &cdx.Metadata{Component: &cdx.Component{Name: "foo", Version: "1.2.3"}}
		files := &mock.FilesMock{}

		err := writeBOM(bom, "bom-helm.xml", files)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not schema-conformant")
		exists, _ := files.FileExists("bom-helm.xml")
		assert.False(t, exists, "a non-conformant BOM must not be written to disk")
	})

	t.Run("write failure is surfaced", func(t *testing.T) {
		bom := newChartBOM(&versioning.Metadata{Name: "foo", Version: "1.2.3"}, "chart", nil, &mock.FilesMock{})
		files := &mock.FilesMock{FileWriteError: fmt.Errorf("disk full")}

		err := writeBOM(bom, "bom-helm.xml", files)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "disk full")
	})
}

func TestRunHelmValidateChartBOM(t *testing.T) {
	// A fully conformant BOM built by newChartBOM, serialized, is the baseline.
	conformant := func() []byte {
		bom := newChartBOM(&versioning.Metadata{Name: "foo", Version: "1.2.3"}, "chart", nil, &mock.FilesMock{})
		var buf bytes.Buffer
		enc := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatXML)
		require.NoError(t, enc.Encode(bom))
		return buf.Bytes()
	}

	t.Run("accepts a conformant BOM", func(t *testing.T) {
		assert.NoError(t, validateChartBOM(conformant()))
	})

	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{"missing serialNumber", func(s string) string {
			return regexp.MustCompile(` serialNumber="[^"]*"`).ReplaceAllString(s, "")
		}, "serialNumber"},
		{"wrong xmlns", func(s string) string {
			return strings.Replace(s, "/bom/1.4", "/bom/1.5", 1)
		}, "unexpected xmlns"},
		{"missing timestamp", func(s string) string {
			return regexp.MustCompile(`<timestamp>[^<]*</timestamp>`).ReplaceAllString(s, "")
		}, "timestamp is required"},
		{"malformed root purl", func(s string) string {
			return strings.Replace(s, "<purl>pkg:helm/foo@1.2.3</purl>", "<purl></purl>", 1)
		}, "purl"},
		{"not XML", func(s string) string { return "not xml" }, "failed to parse"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateChartBOM([]byte(test.mutate(string(conformant()))))
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestRunHelmGenerateChartSBOM(t *testing.T) {
	const validChart = "apiVersion: v2\nname: foo\nversion: 1.2.3\nappVersion: 4.5.6\n"

	t.Run("valid chart produces a valid BOM with a helm PURL", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(validChart))

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		assert.Contains(t, string(content), "pkg:helm/foo@1.2.3")
		// Gateway-required fields must be serialized into the XML.
		assert.Contains(t, string(content), "serialNumber=\"urn:uuid:")
		assert.Contains(t, string(content), "<timestamp>")
		assert.Contains(t, string(content), "bom-ref=\"pkg:helm/foo@1.2.3\"")
		// The BOM must be emitted in CycloneDX schema version 1.4.
		assert.Contains(t, string(content), `xmlns="http://cyclonedx.org/schema/bom/1.4"`)
	})

	t.Run("sub-chart dependencies become components", func(t *testing.T) {
		chartWithDeps := "apiVersion: v2\nname: foo\nversion: 1.2.3\n" +
			"dependencies:\n" +
			"  - name: common\n    version: 2.0.0\n    repository: https://charts.example.com\n" +
			"  - name: postgresql\n    version: 12.1.0\n    repository: https://charts.example.com\n"
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(chartWithDeps))

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		// The chart itself is the root component.
		assert.Contains(t, string(content), "pkg:helm/foo@1.2.3")
		// Each sub-chart dependency is emitted as its own component.
		assert.Contains(t, string(content), "pkg:helm/common@2.0.0")
		assert.Contains(t, string(content), "pkg:helm/postgresql@12.1.0")
	})

	t.Run("Chart.lock resolved versions override Chart.yaml ranges", func(t *testing.T) {
		// Chart.yaml carries version ranges; Chart.lock pins the resolved
		// versions. The lock must win so the SBOM reflects what is packaged.
		chartWithRanges := "apiVersion: v2\nname: foo\nversion: 1.2.3\n" +
			"dependencies:\n" +
			"  - name: common\n    version: ^2.0.0\n    repository: https://charts.example.com\n" +
			"  - name: postgresql\n    version: ~12.0.0\n    repository: https://charts.example.com\n"
		chartLock := "dependencies:\n" +
			"  - name: common\n    version: 2.3.1\n    repository: https://charts.example.com\n" +
			"  - name: postgresql\n    version: 12.1.9\n    repository: https://charts.example.com\n"
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(chartWithRanges))
		files.AddFile("chart/Chart.lock", []byte(chartLock))

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		// Resolved versions from Chart.lock win.
		assert.Contains(t, string(content), "pkg:helm/common@2.3.1")
		assert.Contains(t, string(content), "pkg:helm/postgresql@12.1.9")
		// The unresolved Chart.yaml ranges must not appear.
		assert.NotContains(t, string(content), "pkg:helm/common@^2.0.0")
		assert.NotContains(t, string(content), "pkg:helm/postgresql@~12.0.0")
	})

	t.Run("malformed Chart.lock falls back to Chart.yaml dependencies", func(t *testing.T) {
		chartWithDeps := "apiVersion: v2\nname: foo\nversion: 1.2.3\n" +
			"dependencies:\n" +
			"  - name: common\n    version: 2.0.0\n    repository: https://charts.example.com\n"
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(chartWithDeps))
		files.AddFile("chart/Chart.lock", []byte("dependencies: : broken"))

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.NoError(t, err, "an unparseable Chart.lock must not fail SBOM generation")
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		// Falls back to the Chart.yaml dependency version.
		assert.Contains(t, string(content), "pkg:helm/common@2.0.0")
	})

	t.Run("empty Chart.lock falls back to Chart.yaml dependencies", func(t *testing.T) {
		chartWithDeps := "apiVersion: v2\nname: foo\nversion: 1.2.3\n" +
			"dependencies:\n" +
			"  - name: common\n    version: 2.0.0\n    repository: https://charts.example.com\n"
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(chartWithDeps))
		files.AddFile("chart/Chart.lock", []byte("dependencies: []\n"))

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		// An empty lock dependency list must not suppress the Chart.yaml deps.
		assert.Contains(t, string(content), "pkg:helm/common@2.0.0")
	})

	t.Run("referenced images become container components", func(t *testing.T) {
		const digest = "sha256:0123456789012345678901234567890123456789012345678901234567890123"
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(validChart))
		images := []string{"registry.example.com/foo:1.0.0", "registry.example.com/bar@" + digest}

		err := GenerateChartSBOM("chart", "bom-helm.xml", images, files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		// The chart itself is still the root component.
		assert.Contains(t, string(content), "pkg:helm/foo@1.2.3")
		// Each referenced image is a container component with a pkg:oci PURL.
		assert.Contains(t, string(content), `type="container"`)
		assert.Contains(t, string(content), "pkg:oci/foo@1.0.0")
		assert.Contains(t, string(content), "pkg:oci/bar@"+digest)
	})

	t.Run("unparseable image is skipped, valid images still included", func(t *testing.T) {
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(validChart))
		images := []string{"bad name", "registry.example.com/foo:1.0.0"}

		err := GenerateChartSBOM("chart", "bom-helm.xml", images, files)

		require.NoError(t, err, "an unparseable image must not fail SBOM generation")
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))
		// The valid image is still emitted; the bad one is silently skipped.
		assert.Contains(t, string(content), "pkg:oci/foo@1.0.0")
		assert.NotContains(t, string(content), "bad name")
	})

	t.Run("deps and images combine into the expected component set", func(t *testing.T) {
		chartWithDeps := "apiVersion: v2\nname: foo\nversion: 1.2.3\n" +
			"dependencies:\n" +
			"  - name: common\n    version: 2.0.0\n    repository: https://charts.example.com\n" +
			"  - name: postgresql\n    version: 12.1.0\n    repository: https://charts.example.com\n"
		files := &mock.FilesMock{}
		files.AddFile("chart/Chart.yaml", []byte(chartWithDeps))
		images := []string{"registry.example.com/foo:1.0.0", "registry.example.com/baz:2.0.0"}

		err := GenerateChartSBOM("chart", "bom-helm.xml", images, files)

		require.NoError(t, err)
		content, err := files.FileRead("bom-helm.xml")
		require.NoError(t, err)
		assert.NoError(t, piperutils.ValidateBOM(content))

		// Exactly 2 dep + 2 image components (the chart itself is the root
		// metadata component and is not counted here).
		var bom piperutils.Bom
		require.NoError(t, xml.Unmarshal(content, &bom))
		assert.Len(t, bom.Components, 4, "expected 2 sub-chart deps + 2 images")
		assert.Contains(t, string(content), "pkg:helm/foo@1.2.3")
		assert.Contains(t, string(content), "pkg:helm/common@2.0.0")
		assert.Contains(t, string(content), "pkg:helm/postgresql@12.1.0")
		assert.Contains(t, string(content), "pkg:oci/foo@1.0.0")
		assert.Contains(t, string(content), "pkg:oci/baz@2.0.0")
	})

	t.Run("chart metadata read failure is propagated", func(t *testing.T) {
		files := &mock.FilesMock{} // no Chart.yaml

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read chart/Chart.yaml")
	})

	t.Run("output write failure is surfaced", func(t *testing.T) {
		files := &mock.FilesMock{FileWriteError: fmt.Errorf("disk full")}
		files.AddFile("chart/Chart.yaml", []byte(validChart))

		err := GenerateChartSBOM("chart", "bom-helm.xml", nil, files)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "disk full")
	})
}
