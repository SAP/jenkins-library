package cnbutils

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/linux"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

var cycloneDxXML = syft.FormatByID(syft.CycloneDxXMLFormatID)

func MergeSBOMFiles(pattern, output, img, dockerConfigFile string, utils BuildUtils) error {
	if dockerConfigFile != "" {
		os.Setenv("DOCKER_CONFIG", filepath.Dir(dockerConfigFile))
		defer os.Unsetenv("DOCKER_CONFIG")
	}

	log.Entry().Debugf("reading remote image %s", img)
	remoteImage, err := utils.GetRemoteImageInfo(img)
	if err != nil {
		return err
	}

	imgConfig, err := remoteImage.ConfigFile()
	if err != nil {
		return err
	}

	layerSHA, exists := imgConfig.Config.Labels["io.buildpacks.base.sbom"]
	var bom *sbom.SBOM
	if exists {
		log.Entry().Debug("found SBOM layer")
		bom, err = readBOMFromLayer(remoteImage, layerSHA)
		if err != nil {
			return err
		}
		log.Entry().Debugf("initial source.ImageMetadata: %#v", bom.Source.ImageMetadata)
	} else {
		log.Entry().Debug("SBOM layer not found, creating a new one")
		bom = &sbom.SBOM{
			Artifacts: sbom.Artifacts{
				PackageCatalog: pkg.NewCatalog(),
				LinuxDistribution: &linux.Release{
					Name:    imgConfig.OS,
					Version: imgConfig.OSVersion,
					Variant: imgConfig.Variant,
				},
			},
		}
	}
	imageMetaData, err := extractImageMetaData(remoteImage)
	if err != nil {
		return err
	}
	bom.Source.ImageMetadata = *imageMetaData
	bom.Source.ImageMetadata.UserInput = img
	log.Entry().Debugf("updated source.ImageMetadata: %#v", bom.Source.ImageMetadata)

	log.Entry().Debugf("search for sbom file using the pattern %s", pattern)
	syftFiles, err := utils.Glob(pattern)
	if err != nil {
		return err
	}

	for _, syftFile := range syftFiles {
		log.Entry().Debugf("reading Syft SBOM file %q", syftFile)
		f, err := utils.Open(syftFile)
		if err != nil {
			return err
		}
		defer f.Close()

		bill, _, err := syft.Decode(f)
		if err != nil {
			return err
		}

		for p := range bill.Artifacts.PackageCatalog.Enumerate() {
			bom.Artifacts.PackageCatalog.Add(p)
		}
	}

	//TODO: testability
	outFile, err := filepath.Abs(output)
	if err != nil {
		return err
	}

	out, err := utils.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Printf("*** saving %q ***", outFile)
	log.Entry().Debugf("saving CycloneDX SBOM file to %q", outFile)
	err = cycloneDxXML.Encode(out, *bom)
	if err != nil {
		return err
	}

	return nil
}

func readBOMFromLayer(img v1.Image, layerDiffSHA string) (*sbom.SBOM, error) {
	layerDiffDigest, err := v1.NewHash(layerDiffSHA)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse layer sha %q", layerDiffSHA)
	}

	log.Entry().Debugf("looking for the layer %q", layerDiffDigest.String())

	sbomLayer, err := img.LayerByDiffID(layerDiffDigest)
	if err != nil {
		return nil, errors.Wrapf(err, "failed get layer %q", layerDiffDigest)
	}

	rc, err := sbomLayer.Uncompressed()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get uncompressed reader")
	}

	tr := tar.NewReader(rc)
	sbomRegex := regexp.MustCompile(`cnb/sbom/[a-z0-9]+\.syft\.json`)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to read tar content")
		}

		log.Entry().Debugf("checking SBOM layer file %q", hdr.Name)
		if sbomRegex.Match([]byte(hdr.Name)) {
			log.Entry().Debugf("file %q matches the regex", hdr.Name)
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, tr)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read SBOM file from the layer")
			}

			bom, _, err := syft.Decode(buf)
			return bom, errors.Wrap(err, "failed to decode SBOM file")
		}
	}

	return nil, errors.New("no sbom file found")
}

func extractImageMetaData(img v1.Image) (*source.ImageMetadata, error) {
	imageDigest, err := img.Digest()
	if err != nil {
		return nil, err
	}

	imageMediaType, err := img.MediaType()
	if err != nil {
		return nil, err
	}

	imageSize, err := img.Size()
	if err != nil {
		return nil, err
	}

	imageRawManifest, err := img.RawManifest()
	if err != nil {
		return nil, err
	}

	imageRawConfig, err := img.RawConfigFile()
	if err != nil {
		return nil, err
	}

	return &source.ImageMetadata{
		ID:          imageDigest.String(),
		Size:        imageSize,
		MediaType:   string(imageMediaType),
		RawConfig:   imageRawConfig,
		RawManifest: imageRawManifest,
	}, nil
}
