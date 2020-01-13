package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/protecode"
)

func protecodeExecuteScan(myProtecodeExecuteScanOptions protecodeExecuteScanOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	//create client for sending api request
	client := createClient(myProtecodeExecuteScanOptions)

	return runProtecodeScan(myProtecodeExecuteScanOptions, &c, client)
}

func runProtecodeScan(myProtecodeExecuteScanOptions protecodeExecuteScanOptions, command execRunner, client protecode.Protecode) error {

	//load existing product by filename
	productId, err := client.LoadExistingProduct(myProtecodeExecuteScanOptions.ProtecodeGroup, myProtecodeExecuteScanOptions.FilePath, myProtecodeExecuteScanOptions.ReuseExisting)
	if err != nil {
		return err
	}

	// check if no existing is found or reuse existing is false
	productId, err = uploadScanOrDeclareFetch(myProtecodeExecuteScanOptions, productId, client)
	if err != nil {
		return err
	}
	if(productId < 0) {
		return errors.New("Protecode scan failed, the product id is below zero")
	}
	//pollForResult
	result, err := client.PollForResult(productId, myProtecodeExecuteScanOptions.Verbose)
	if err != nil {
		return err
	}
	//check if result is ok else notify
	if len(result.Status) > 0 || result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
		//Notify.error(this, "Protecode scan failed, please check the log and protecode backend for more details.")
		return errors.New("Protecode scan failed, please check the log and protecode backend for more details.")
	}

	//loadReport
	resp, err := client.LoadReport(myProtecodeExecuteScanOptions.ReportFileName, productId)
	if err != nil {
		return err
	}
	//save report to filesystem
	err = writeReportToFile(*resp, myProtecodeExecuteScanOptions.ReportFileName)
	if err != nil {
		return err
	}

	//count vulnerabilities
	m := client.ParseResultForInflux(result, myProtecodeExecuteScanOptions.ProtecodeExcludeCVEs)

	//write result to the filesysten
	err = writeResultAsJSONToFile(m, "VulnResult.json", fileWriter)
	if err != nil {
		return err
	}

	//clean scan from server
	err = client.DeleteScan(myProtecodeExecuteScanOptions.CleanupMode, productId)
	if err != nil {
		return err
	}

	return nil
}

func createClient(config protecodeExecuteScanOptions) protecode.Protecode {

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

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productId int, client protecode.Protecode) (int, error) {

	// check if no existing is found or reuse existing is false
	if productId == 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			result, err := client.DeclareFetchUrl(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			if err != nil {
				return -1, err
			}
			productId = result.ProductId
			
		} else {
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			result, err := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FilePath)
			if err != nil {
				return -1, err
			}
			productId = result.ProductId
		}
	}

	return productId, nil
}

func writeResultAsJSONToFile(m map[string]int, filename string, writeFunc func(f string, b []byte, p os.FileMode) error) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return writeFunc(filename, b, 644)
}

func writeReportToFile(resp io.ReadCloser, reportFileName string) error {
	f, err := os.Create(reportFileName)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}

func fileWriter(filename string, b []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, b, perm)
}
