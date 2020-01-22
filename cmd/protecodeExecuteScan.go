package cmd

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/protecode"
	dchClient "github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
)

func protecodeExecuteScan(config protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment, influx *protecodeExecuteScanInflux) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	return runProtecodeScan(&config, cpEnvironment, influx)
}

func runProtecodeScan(config *protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment, influx *protecodeExecuteScanInflux) error {

	//create client for sending api request
	client := createClient(config)

	err := handleDockerCredentials(config)
	if err != nil {
		log.Entry().Fatalf("Exception during the handling of the credentials %v", err)
	}
	fileName, err := getDockerImage(config, cpEnvironment)
	if err != nil {
		log.Entry().Fatalf("Exception during getting the image %v", err)
	}
	err = cleanupDockerCredentials(config)
	if err != nil {
		log.Entry().Fatalf("Exception during the cleanup of the credentials %v", err)
	}
	parsedResult, productId, err := executeProtecodeScan(client, config, fileName, writeReportToFile)
	if err != nil {
		log.Entry().Fatalf("Exception during the execute of the scan %v", err)
	}

	setInfluxData(influx, parsedResult)

	setCommonPipelineEnvironmentData(cpEnvironment, parsedResult, productId)

	log.Entry().Debugf("Cleanup tar archive")
	err = os.Remove(config.FilePath)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to delete tar source code")
	}

	return nil
}

func handleDockerCredentials(config *protecodeExecuteScanOptions) error {

	if len(config.DockerUser) > 0 && len(config.DockerPassword) > 0 {
		//create config file
		f, err := os.Create("~/.docker/config.json")
		if err == nil {
			defer f.Close()
			_, err = f.WriteString(`{\n"credsStore": "secretservice"\n}`)
			if err != nil {
				log.Entry().Fatalf("Exception during writing the credentials store configuration %v", err)
			}

			p := dchClient.NewShellProgramFunc("docker-credential-secretservice")

			c := &credentials.Credentials{
				ServerURL: config.DockerRegistryURL,
				Username:  config.DockerUser,
				Secret:    config.DockerPassword,
			}

			//add credentials
			if err := dchClient.Store(p, c); err != nil {
				if err := dchClient.Erase(p, config.DockerRegistryURL); err != nil {
					log.Entry().Fatalf("Exception during erase the credentials %v", err)
				}
				log.Entry().Fatalf("Exception during store the credentials %v", err)
			}
		}
	}

	return nil
}

func cleanupDockerCredentials(config *protecodeExecuteScanOptions) error {
	if len(config.DockerUser) > 0 && len(config.DockerPassword) > 0 {
		p := dchClient.NewShellProgramFunc("docker-credential-secretservice")

		if err := dchClient.Erase(p, config.DockerRegistryURL); err != nil {
			log.Entry().Fatalf("Exception during erase the credentials %v", err)
		}
	}

	return nil
}

func getDockerImage(config *protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment) (string, error) {

	cacheImagePath := "./cache/protecodeImage"
	completeURL, err := getUrlAndFileNameFromDockerImage(config)
	if err != nil {
		log.Entry().Fatalf("Exception during get url creation for get the docker image %v", err)
	}

	image, err := pkgutil.GetImage(completeURL, config.IncludeLayers, cacheImagePath)
	if err != nil {
		log.Entry().Fatalf("Exception during get image %v", err)
	}
	fileName := fmt.Sprintf("%v.tar", strings.ReplaceAll(config.ScanImage, "/", "_"))

	//tar folder
	cacheTarPath := "./cache"
	tarFileName := filepath.Join(cacheTarPath, fileName)
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create archive of docker image")
	}
	defer tarFile.Close()
	err = tarImageFolder(cacheImagePath, tarFile)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to create tar archive of docker image")
	}

	if len(config.FilePath) <= 0 {
		(*config).FilePath = fmt.Sprintf("./%v", filepath.Join("./", tarFileName))
		if len(config.FilePath) <= 0 {
			log.Entry().Fatalf("Protecode scan failed, there is no file path configured  : %v (filename:%v, PSPath: %v)", config.FilePath, fileName, image.FSPath)
		}
	}
	return fileName, nil
}

func getUrlAndFileNameFromDockerImage(config *protecodeExecuteScanOptions) (string, error) {

	completeUrl := config.ScanImage

	if len(config.DockerRegistryURL) > 0 {
		if strings.HasSuffix(config.DockerRegistryURL, "/") {
			completeUrl = fmt.Sprintf("remote://%v%v", config.DockerRegistryURL, config.ScanImage)
		} else {
			completeUrl = fmt.Sprintf("remote://%v/%v", config.DockerRegistryURL, config.ScanImage)
		}
	}

	if len(completeUrl) <= 0 {
		return completeUrl, errors.New("Protecode scan failed, there is no scan image configured")
	}

	return completeUrl, nil
}

func tarImageFolder(source string, tarFile io.Writer) error {
	archive := tar.NewWriter(tarFile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		if err := writeToTar(path, source, archive, info); err != nil {
			return err
		}
		return nil
	})

	err := archive.Close()

	return err
}

func writeToTar(fileDir string,
	sourceBase string,
	archive *tar.Writer,
	info os.FileInfo) error {

	file, err := os.Open(fileDir)
	if err != nil {
		return err
	}
	defer file.Close()

	// relative paths are used to preserve the directory paths in each file path
	relativePath, err := filepath.Rel(sourceBase, fileDir)

	tarHeader, err := tar.FileInfoHeader(info, relativePath)
	if err != nil {
		return err
	}

	err = archive.WriteHeader(tarHeader)
	if err != nil {
		return err
	}
	_, err = io.Copy(archive, file)
	if err != nil {
		return err
	}
	return nil
}

func executeProtecodeScan(client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) (map[string]int, int, error) {

	var parsedResult map[string]int = make(map[string]int)
	//load existing product by filename
	productId, err := client.LoadExistingProduct(config.ProtecodeGroup, config.FilePath, config.ReuseExisting)
	if err != nil {
		return parsedResult, productId, err
	}
	// check if no existing is found or reuse existing is false
	productId, err = uploadScanOrDeclareFetch(*config, productId, client, fileName)
	if err != nil {
		return parsedResult, productId, err
	}
	if productId <= 0 {
		return parsedResult, productId, errors.New("Protecode scan failed, the product id is not valid (product id <= zero)")
	}
	//pollForResult
	result, err := client.PollForResult(productId, config.Verbose)
	if err != nil {
		return parsedResult, productId, err
	}
	//check if result is ok else notify
	if len(result.Status) > 0 && result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
		return parsedResult, productId, errors.New("Protecode scan failed, please check the log and protecode backend for more details.")
	}
	//loadReport
	resp, err := client.LoadReport(config.ReportFileName, productId)
	if err != nil {
		return parsedResult, productId, err
	}
	//save report to filesystem
	err = writeReportToFile(*resp, config.ReportFileName)
	if err != nil {
		return parsedResult, productId, err
	}
	//clean scan from server
	err = client.DeleteScan(config.CleanupMode, productId)
	if err != nil {
		return parsedResult, productId, err
	}
	//count vulnerabilities
	parsedResult = client.ParseResultForInflux(result, config.ProtecodeExcludeCVEs)

	return parsedResult, productId, nil
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {

	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["minor_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["major_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["vulnerabilities"])
}

func setCommonPipelineEnvironmentData(cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment, result map[string]int, productId int) {

	cpEnvironment.appContainerProperties.protecodeProductID = fmt.Sprintf("%v", productId)
	cpEnvironment.appContainerProperties.protecodeCount = fmt.Sprintf("%v", result["count"])
	cpEnvironment.appContainerProperties.cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	cpEnvironment.appContainerProperties.cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	cpEnvironment.appContainerProperties.excluded_vulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	cpEnvironment.appContainerProperties.triaged_vulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	cpEnvironment.appContainerProperties.historical_vulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
}

func createClient(config *protecodeExecuteScanOptions) protecode.Protecode {

	var duration time.Duration = time.Duration(10 * 60)

	if len(config.ProtecodeTimeoutMinutes) > 0 {
		s, _ := strconv.ParseInt(config.ProtecodeTimeoutMinutes, 10, 64)
		duration = time.Duration(s * 60)
	}

	pc := protecode.Protecode{}

	protecodeOptions := protecode.ProtecodeOptions{
		ServerURL: config.ProtecodeServerURL,
		Logger:    log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode"),
		Duration:  duration,
		Username:  config.User,
		Password:  config.Password,
	}

	pc.SetOptions(protecodeOptions)

	return pc
}

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productId int, client protecode.Protecode, filaName string) (int, error) {

	// check if no existing is found or reuse existing is false
	if productId <= 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			result, err := client.DeclareFetchUrl(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			if err != nil {
				return -1, err
			}
			productId = result.ProductId

		} else {
			if len(config.FilePath) <= 0 {
				return -1, errors.New(fmt.Sprintf("Protecode scan failed, there is no file path configured for upload : %v", config.FilePath))
			}
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			result, err := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FilePath, filaName)
			if err != nil {
				return -1, err
			}
			productId = result.ProductId
		}
	}

	return productId, nil
}

func writeReportToFile(resp io.ReadCloser, reportFileName string) error {
	f, err := os.Create(reportFileName)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}
