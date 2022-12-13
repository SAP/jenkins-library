package syft

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

func GenerateSBOM(syftDownloadURL, dockerConfigDir string, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender, registryURL string, images []string) error {
	if registryURL == "" {
		return errors.New("syft: regisitry url must not be empty")
	}

	if len(images) == 0 {
		return errors.New("syft: no images provided")
	}

	execRunner.AppendEnv([]string{fmt.Sprintf("DOCKER_CONFIG=%s", dockerConfigDir)})

	tmpDir, err := fileUtils.TempDir("", "syft")
	if err != nil {
		return err
	}
	syftFile := filepath.Join(tmpDir, "syft")

	err = install(syftDownloadURL, syftFile, fileUtils, httpClient)
	if err != nil {
		return errors.Wrap(err, "failed to install syft")
	}

	for index, image := range images {
		if image == "" {
			return errors.New("syft: image name must not be empty")
		}
		// TrimPrefix needed as syft needs containerRegistry name only
		err = execRunner.RunExecutable(syftFile, "packages", fmt.Sprintf("registry:%s/%s", strings.TrimPrefix(registryURL, "https://"), image), "-o", "cyclonedx-xml", "--file", fmt.Sprintf("bom-docker-%v.xml", index), "-q")
		if err != nil {
			return fmt.Errorf("failed to generate SBOM: %w", err)
		}
	}

	return nil
}

func install(syftDownloadURL, dest string, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {
	response, err := httpClient.SendRequest(http.MethodGet, syftDownloadURL, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to download syft binary: %w", err)
	}
	defer response.Body.Close()

	err = extractSyft(response.Body, dest, fileUtils)
	if err != nil {
		return errors.Wrap(err, "failed to extract syft binary")
	}

	err = fileUtils.Chmod(dest, 0755)
	if err != nil {
		return err
	}

	return nil
}

func extractSyft(archive io.Reader, dest string, fileUtils piperutils.FileUtils) error {
	zr, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	fileFound := false
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return errors.Wrap(err, "failed to read archive")
		}

		if filepath.Base(f.Name) == "syft" {
			fileFound = true

			df, err := fileUtils.Create(dest)
			if err != nil {
				return errors.Wrapf(err, "failed to create file %q", dest)
			}

			size, err := io.Copy(df, tr)
			if err != nil {
				return err
			}

			err = df.Close()
			if err != nil {
				return err
			}

			if size != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", size, dest, f.Size)
			}
		}
	}

	if !fileFound {
		return errors.New("no file with the name 'syft' was found")
	}

	return nil
}
