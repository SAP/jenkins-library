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

func executeProtecodeScan(myExecuteProtecodeScanOptions executeProtecodeScanOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	runProtecodeScan(myExecuteProtecodeScanOptions, &c)
	return nil
}

func runProtecodeScan(myExecuteProtecodeScanOptions executeProtecodeScanOptions, command execRunner) error {

	//create client for sending api request
	client, dur := createClient(myExecuteProtecodeScanOptions)

	//load existing product by filename
	productId, err := client.LoadExistingProduct(myExecuteProtecodeScanOptions.ProtecodeGroup, myExecuteProtecodeScanOptions.FilePath, myExecuteProtecodeScanOptions.ReuseExisting)
	if err != nil {
		return err
	}

	// check if no existing is found or reuse existing is false
	productId, err = uploadScanOrDeclareFetch(myExecuteProtecodeScanOptions, productId, client)
	if err != nil {
		return err
	}
	//pollForResult
	result, err := client.PollForResult(productId, myExecuteProtecodeScanOptions.Verbose, dur)
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
	resp, err := client.LoadReport(myExecuteProtecodeScanOptions.ReportFileName, productId)
	if err != nil {
		return err
	}
	//save to filesystem
	writeReportToFile(*resp, myExecuteProtecodeScanOptions.ReportFileName)

	//count vulnerabilities
	m := client.ParseResultForInflux(result, myExecuteProtecodeScanOptions.ProtecodeExcludeCVEs)

	err = writeResultAsJSONToFile(m, "VulnResult.json", fileWriter)
	if err != nil {
		return err
	}

	//clean scan from server
	if myExecuteProtecodeScanOptions.CleanupMode == "complete" {
		fmt.Printf("Protecode scan successful. Deleting scan from server.")
		client.DeleteScan(myExecuteProtecodeScanOptions.CleanupMode, productId)
	}

	return nil
}

func createClient(config executeProtecodeScanOptions) (*protecode.Protecode, time.Duration) {

	var duration time.Duration = time.Duration(10 * 60)

	if len(config.ProtecodeTimeoutMinutes) > 0 {
		s, _ := strconv.ParseInt(config.ProtecodeTimeoutMinutes, 10, 64)
		duration = time.Duration(s * 60)
	}
	client := protecode.New(config.ProtecodeServerURL, duration, config.User, config.Password)

	return client, duration
}

func uploadScanOrDeclareFetch(config executeProtecodeScanOptions, productId int, client *protecode.Protecode) (int, error) {

	// check if no existing is found or reuse existing is false
	if productId == 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			result, err := client.DeclareFetchUrl(config.CleanupMode, config.ProtecodeGroup, config.FilePath)
			if err != nil {
				return 0, err
			}
			productId = result.ProductId
		} else {
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			result, err := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			if err != nil {
				return 0, err
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

func writeReportToFile(resp io.ReadCloser, reportFileName string) {
	f, err := os.Create(reportFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = io.Copy(f, resp)
	if err != nil {
		panic(err)
	}
}

func fileWriter(filename string, b []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, b, perm)
}
