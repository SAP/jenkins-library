package syft

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type SyftScanner struct {
	syftFile       string
	additionalArgs []string
}

const cyclonedxFormatForSyft = "@1.4"

func GenerateSBOM(syftDownloadURL, dockerConfigDir string, execRunner command.ExecRunner, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender, registryURL string, images []string) error {
	scanner, err := CreateSyftScanner(syftDownloadURL, fileUtils, httpClient)
	if err != nil {
		return err
	}
	return scanner.ScanImages(dockerConfigDir, execRunner, registryURL, images)
}

func CreateSyftScanner(syftDownloadURL string, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) (*SyftScanner, error) {

	tmpDir, err := fileUtils.TempDir("", "syft")
	if err != nil {
		return nil, err
	}
	syftFile := filepath.Join(tmpDir, "syft")

	err = install(syftDownloadURL, syftFile, fileUtils, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to install syft: %w", err)
	}

	return &SyftScanner{syftFile: syftFile}, nil
}

func (s *SyftScanner) AddArgument(arg string) {
	s.additionalArgs = append(s.additionalArgs, arg)
}

func (s *SyftScanner) ScanImages(dockerConfigDir string, execRunner command.ExecRunner, registryURL string, images []string) error {
	if registryURL == "" {
		return errors.New("syft: registry url must not be empty")
	}

	if len(images) == 0 {
		return errors.New("syft: no images provided")
	}

	execRunner.AppendEnv([]string{fmt.Sprintf("DOCKER_CONFIG=%s", dockerConfigDir)})

	for index, image := range images {
		if image == "" {
			return errors.New("syft: image name must not be empty")
		}
		// TrimPrefix needed as syft needs containerRegistry name only
		args := []string{"scan", fmt.Sprintf("registry:%s/%s", strings.TrimPrefix(registryURL, "https://"), image), "-o", fmt.Sprintf("cyclonedx-xml%s=bom-docker-%v.xml", cyclonedxFormatForSyft, index), "-q"}
		args = append(args, s.additionalArgs...)
		err := execRunner.RunExecutable(s.syftFile, args...)
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
		return fmt.Errorf("failed to extract syft binary: %w", err)
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
			return fmt.Errorf("failed to read archive: %w", err)
		}

		if filepath.Base(f.Name) == "syft" {
			fileFound = true

			df, err := fileUtils.Create(dest)
			if err != nil {
				return fmt.Errorf("failed to create file %q: %w", dest, err)
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
