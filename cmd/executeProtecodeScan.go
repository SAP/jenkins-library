package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
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
	productId, err := loadExistingProduct(myExecuteProtecodeScanOptions, client)
	if err != nil {
		return err
	}

	// check if no existing is found or reuse existing is false
	productId, err = uploadScanOrDeclareFetch(myExecuteProtecodeScanOptions, productId)
	if err != nil {
		return err
	}
	//pollForResult
	result, err := pollForResult(myExecuteProtecodeScanOptions, productId, client, dur)
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
	resp, err := loadReport(myExecuteProtecodeScanOptions, productId, client)
	if err != nil {
		return err
	}
	//save to filesystem
	writeReportToFile(*resp, myExecuteProtecodeScanOptions.ReportFileName)

	//count vulnerabilities
	m := protecode.ParseResultForInflux(result, myExecuteProtecodeScanOptions.ProtecodeExcludeCVEs)

	//TODO write JSON
	err = protecode.WriteResultAsJSONToFile(m, "VulnResult.json", fileWriter)
	if err != nil {
		return err
	}

	//clean scan from server
	if myExecuteProtecodeScanOptions.CleanupMode == "complete" {
		fmt.Printf("Protecode scan successful. Deleting scan from server.")
		deleteScan(myExecuteProtecodeScanOptions, productId, client)
	}

	return nil
}

func createClient(config executeProtecodeScanOptions) (piperHttp.Client, time.Duration) {

	var dur time.Duration = time.Duration(10 * 60)

	client := piperHttp.Client{}
	if len(config.ProtecodeTimeoutMinutes) > 0 {
		s, _ := strconv.ParseInt(config.ProtecodeTimeoutMinutes, 10, 64)
		dur = time.Duration(s * 60)
	}
	opts := piperHttp.ClientOptions{dur, config.User, config.Password, ""}
	client.SetOptions(opts)

	return client, dur
}

func loadExistingProduct(config executeProtecodeScanOptions, client piperHttp.Client) (int, error) {
	var productId int = 0

	if config.ReuseExisting {

		response, err := loadExistingProductByFilename(config, client)
		if err != nil {
			return 0, err
		}
		// by definition we will take the first one and trigger rescan
		productId = response.Products[0].ProductId

		fmt.Printf("re-use existing Protecode scan - file: %v, group: %v, productId: %v", myExecuteProtecodeScanOptions.FilePath, myExecuteProtecodeScanOptions.ProtecodeGroup, productId)
	}

	return productId, nil
}

func loadExistingProductByFilename(config executeProtecodeScanOptions, client piperHttp.Client) (*protecode.ProductData, error) {

	protecodeURL, headers, err := getLoadExistiongProductRequestData(config)

	if err != nil {
		return new(protecode.ProductData), err
	}

	return loadExisting(protecodeURL, headers, client)
}

func getLoadExistiongProductRequestData(config executeProtecodeScanOptions) (string, map[string][]string, error) {

	protecodeURL, err := protecode.CreateUrl(config.ProtecodeServerURL, "/api/apps/", fmt.Sprintf("%v/", config.ProtecodeGroup), config.FilePath)
	headers := map[string][]string{
		//change to mimetype
		"acceptType": []string{"APPLICATION_JSON"},
	}

	return protecodeURL, headers, err
}

func loadExisting(protecodeURL string, headers map[string][]string, client piperHttp.Client) (*protecode.ProductData, error) {

	r, err := protecode.SendApiRequest("GET", protecodeURL, headers, client)
	if err != nil {
		return new(protecode.ProductData), err
	}

	return protecode.GetProductData(*r)
}

func pullResult(config executeProtecodeScanOptions, productId int, client piperHttp.Client) (protecode.Result, error) {
	protecodeURL, headers, err := getPullResultRequestData(config, productId)
	if err != nil {
		return *new(protecode.Result), err
	}

	return pullResultData(protecodeURL, headers, client)

}

func pullResultData(protecodeURL string, headers map[string][]string, client piperHttp.Client) (protecode.Result, error) {
	r, err := protecode.SendApiRequest("GET", protecodeURL, headers, client)

	response, err := protecode.GetResultData(*r)

	return response.Result, err
}

func getPullResultRequestData(config executeProtecodeScanOptions, productId int) (string, map[string][]string, error) {
	protecodeURL, err := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/", productId), "")
	headers := map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	}

	return protecodeURL, headers, err
}

func cmdStringUploadScanFile(config executeProtecodeScanOptions) string {
	deleteBinary := (config.CleanupMode == "binary" || config.CleanupMode == "complete")

	cmdString := fmt.Sprintf("curl --insecure -H 'Authorization: Basic %v' -H 'Group: %v' -H 'Delete-Binary: %v' -T %v %v/api/upload/ --write-out '%vstatus=%v'",
		protecode.GetBase64UserPassword(), config.ProtecodeGroup, deleteBinary, config.FilePath, config.ProtecodeServerURL, protecode.DELIMITER, "%{http_code}")

	return cmdString
}

func cmdStringDeclareFetchUrl(config executeProtecodeScanOptions) string {
	deleteBinary := (config.CleanupMode == "binary" || config.CleanupMode == "complete")
	cmdString := fmt.Sprintf("curl -X POST -H 'Authorization: Basic %v' -H 'Group: %v' -H 'Delete-Binary: %v' -H 'Url:%v'  %v/api/fetch/ --write-out '%vstatus=%v'",
		protecode.GetBase64UserPassword(), config.ProtecodeGroup, deleteBinary, config.FetchURL, config.ProtecodeServerURL, protecode.DELIMITER, "%{http_code}")

	return cmdString
}

func uploadScanFile(config executeProtecodeScanOptions) (*protecode.Result, error) {
	deleteBinary := (config.CleanupMode == "binary" || config.CleanupMode == "complete")
	headers := map[string][]string{"Group": []string{""}, "Delete-Binary": []string{fmt.Sprintf("", deleteBinary)}}

	r, err := protecode.UploadScanFile(fmt.Sprintf("%v/api/upload/", config.ProtecodeServerURL), config.FilePath, headers)
	if err != nil {
		log.Entry().WithError(err).Fatalf("error during %v: %v reuqest", method, url)
		return nil, err
	}
	return protecode.GetResult(r, cmdString)

}

func declareFetchUrl(config executeProtecodeScanOptions) protecode.Result {

	cmdString := cmdStringDeclareFetchUrl(config)
	response := protecode.CmdExecGetResult("curl", cmdString)

	return response
}

func uploadScanOrDeclareFetch(config executeProtecodeScanOptions, productId int) (int, error) {

	// check if no existing is found or reuse existing is false
	if productId == 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			result := declareFetchUrl(config)
			productId = result.ProductId
		} else {
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			result, err := uploadScanFile(config)
			if err != nil {
				return 0, err
			}
			productId = result.ProductId
		}
	}

	return productId, nil
}

func pollForResult(config executeProtecodeScanOptions, productId int, client piperHttp.Client, duration time.Duration) (protecode.Result, error) {

	var response protecode.Result
	var err error

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	ticks := duration / 10

	for i := ticks; i > 0; i-- {

		response, err = pullResult(config, productId, client)
		if err != nil {
			ticker.Stop()
			i = 0
			return response, err
		}
		if len(response.Components) > 0 && response.Status != "B" {
			ticker.Stop()
			i = 0
			break
		}

		select {
		case t := <-ticker.C:
			fmt.Printf("Ticker %v", t)
			response, err = pullResult(config, productId, client)
			if err != nil {
				ticker.Stop()
				i = 0
				return response, err
			}
			if len(response.Components) > 0 && response.Status != "B" {
				ticker.Stop()
				i = 0
				break
			}
			if config.Verbose {
				fmt.Printf("Processing status for productId %v", productId)
			}
		}
	}

	if len(response.Components) == 0 && response.Status == "B" {
		response, err = pullResult(config, productId, client)
		if err != nil || len(response.Components) == 0 || response.Status == "B" {
			log.Entry().Fatal("No result for protecode scan")
			return response, err
		}
	}

	return response, nil
}

func loadReport(config executeProtecodeScanOptions, productId int, client piperHttp.Client) (*io.ReadCloser, error) {

	protecodeURL, err := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/pdf-report", productId), "")
	if err != nil {
		return nil, err
	}
	headers := map[string][]string{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Pragma":        []string{"no-cache"},
		"Outputfile":    []string{config.ReportFileName},
	}

	r, e := protecode.SendApiRequest("GET", protecodeURL, headers, client)
	if e != nil {
		return r, e
	}
	return r, nil
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

func deleteScan(config executeProtecodeScanOptions, productId int, client piperHttp.Client) error {

	switch config.CleanupMode {
	case "none":
	case "binary":
		return nil
	case "complete":
		protecodeURL, err := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/", productId), "")
		if err != nil {
			return err
		}
		headers := map[string][]string{}

		_, err = protecode.SendApiRequest("DELETE", protecodeURL, headers, client)
		if err != nil {
			return err
		}
		break
	default:
		log.Entry().Fatalf("Unknown cleanup mode %v", config.CleanupMode)
	}

	return nil
}

func fileWriter(filename string, b []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, b, perm)
}
