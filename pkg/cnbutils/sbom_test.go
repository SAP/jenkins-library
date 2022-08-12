package cnbutils_test

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/mock"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
)

type uncompressedLayer struct {
	diffID    v1.Hash
	mediaType types.MediaType
	content   []byte
}

// DiffID implements partial.UncompressedLayer
func (ul *uncompressedLayer) DiffID() (v1.Hash, error) {
	return ul.diffID, nil
}

// Uncompressed implements partial.UncompressedLayer
func (ul *uncompressedLayer) Uncompressed() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer(ul.content)), nil
}

// MediaType returns the media type of the layer
func (ul *uncompressedLayer) MediaType() (types.MediaType, error) {
	return ul.mediaType, nil
}

func fakeLayer(path string, content []byte) (v1.Layer, error) {
	var b bytes.Buffer
	hasher := sha256.New()
	mw := io.MultiWriter(&b, hasher)

	// Write a single file with a random name and random contents.
	tw := tar.NewWriter(mw)
	if err := tw.WriteHeader(&tar.Header{
		Name:     path,
		Size:     int64(len(content)),
		Typeflag: tar.TypeRegA,
	}); err != nil {
		return nil, err
	}

	if _, err := io.WriteString(tw, string(content)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	h := v1.Hash{
		Algorithm: "sha256",
		Hex:       hex.EncodeToString(hasher.Sum(make([]byte, 0, hasher.Size()))),
	}

	return partial.UncompressedToLayer(&uncompressedLayer{
		mediaType: types.DockerLayer,
		diffID:    h,
		content:   b.Bytes(),
	})
}

func TestMergeSBOMFiles(t *testing.T) {
	var mockUtils = &cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		DownloadMock:   &mock.DownloadMock{},
	}

	fakeImg := &fake.FakeImage{}
	fakeImg.ConfigFileReturns(&v1.ConfigFile{
		Config: v1.Config{Labels: map[string]string{"io.buildpacks.base.sbom": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"}},
	}, nil)

	baseSBOM, err := os.ReadFile("testdata/base.syft.json")
	assert.NoError(t, err)
	fLayer, err := fakeLayer("cnb/sbom/test.syft.json", baseSBOM)
	assert.NoError(t, err)
	fakeImg.LayerByDiffIDReturns(fLayer, nil)

	mockUtils.ReturnImage = fakeImg
	mockUtils.RemoteImageInfo = fakeImg

	//TODO: not found in test
	sbom1, err := os.ReadFile("testdata/sbom1.xml")
	assert.NoError(t, err)
	mockUtils.FilesMock.AddFile("/layer/1/sbom.xml", sbom1)
	sbom2, err := os.ReadFile("testdata/sbom2.xml")
	assert.NoError(t, err)
	mockUtils.FilesMock.AddFile("/layer/2/sbom.xml", sbom2)

	err = cnbutils.MergeSBOMFiles("/layer/**/sbom.xml", "sbom.xml", "imageName", "", mockUtils)
	assert.NoError(t, err)

	exists, err := mockUtils.FilesMock.FileExists("sbom.xml")
	assert.NoError(t, err)
	assert.True(t, exists)
}
