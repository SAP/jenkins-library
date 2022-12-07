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

const syftArchiveName = "syft.tar.gz"

func GenerateSBOM(syftDownloadURL, dockerConfigDir string, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender, registryURL string, images []string) error {
	execRunner.AppendEnv([]string{"DOCKER_CONFIG", dockerConfigDir})

	tmpDir, err := fileUtils.TempDir("", "syft")
	if err != nil {
		return err
	}
	syftFile := filepath.Join(tmpDir, "syft")

	err = install(syftDownloadURL, syftFile, fileUtils, httpClient)
	if err != nil {
		return err
	}

	for index, image := range images {
		// TrimPrefix needed as syft needs containerRegistry name only
		err = execRunner.RunExecutable(syftFile, "packages", fmt.Sprintf("%s/%s", strings.TrimPrefix(registryURL, "https://"), image), "-o", "cyclonedx-xml", "--file", fmt.Sprintf("bom-docker-%v.xml", index))
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

	archiveFile, err := fileUtils.Create(syftArchiveName)
	if err != nil {
		return errors.Wrap(err, "failed to create syft archive file")
	}
	defer archiveFile.Close()

	_, err = piperutils.CopyData(archiveFile, response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write syft archive to disk")
	}

	err = extractSyft(archiveFile, dest, fileUtils)
	if err != nil {
		return errors.Wrap(err, "failed to extract syft binary")
	}

	err = fileUtils.Chmod(dest, 0755)
	if err != nil {
		return err
	}

	return nil
}

func extractSyft(archive io.ReadWriteCloser, dest string, fileUtils piperutils.FileUtils) error {
	zr, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}

	tr := tar.NewReader(zr)

	fileFound := false
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if filepath.Base(f.Name) == "syft" {
			fileFound = true

			sf, err := fileUtils.Create(dest)
			if err != nil {
				return err
			}

			size, err := io.Copy(sf, tr)
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
