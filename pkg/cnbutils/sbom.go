package cnbutils

import (
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/linux"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
)

var cycloneDxXML = syft.FormatByID(syft.CycloneDxXMLFormatID)

func MergeSBOMFiles(pattern, output, img, dockerConfigFile string, utils BuildUtils) error {
	if dockerConfigFile != "" {
		os.Setenv("DOCKER_CONFIG", filepath.Dir(dockerConfigFile))
		defer os.Unsetenv("DOCKER_CONFIG")
	}

	log.Entry().Debugf("search for sbom file using the pattern %s", pattern)
	syftFiles, err := utils.Glob(pattern)
	if err != nil {
		return err
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

	bom := &sbom.SBOM{
		Artifacts: sbom.Artifacts{
			PackageCatalog: pkg.NewCatalog(),
			LinuxDistribution: &linux.Release{
				Name:    imgConfig.OS,
				Version: imgConfig.OSVersion,
				Variant: imgConfig.Variant,
			},
		},
		Source: source.Metadata{
			ImageMetadata: source.NewImageMetadata(image.NewImage(remoteImage, "/tmp"), ""),
		},
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

	outFile, err := filepath.Abs(output)
	if err != nil {
		return err
	}

	out, err := utils.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	log.Entry().Debugf("saving CycloneDX SBOM file to %q", outFile)
	err = cycloneDxXML.Encode(out, *bom)
	if err != nil {
		return err
	}

	return nil
}
