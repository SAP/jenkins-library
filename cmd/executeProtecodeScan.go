package cmd

import (
	"encoding/json"
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

func runProtecodeScan(myExecuteProtecodeScanOptions executeProtecodeScanOptions, command execRunner) {

	//create client for sending api request
	client, dur := createClient(myExecuteProtecodeScanOptions)

	//load existing product by filename
	productId := loadExistingProduct(myExecuteProtecodeScanOptions, client)

	// check if no existing is found or reuse existing is false
	productId = uploadScanOrDeclareFetch(myExecuteProtecodeScanOptions, productId)

	//pollForResult
	result := pollForResult(myExecuteProtecodeScanOptions, productId, client, dur)

	//check if result is ok else notify
	if len(result.Status) > 0 || result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
		os.Exit(1)
		//Notify.error(this, "Protecode scan failed, please check the log and protecode backend for more details.")
	}

	//loadReport
	resp := loadReport(myExecuteProtecodeScanOptions, productId, client)

	//save to filesystem
	writeHttpResponseToFile(*resp)

	//count vulnerabilities
	m := protecode.ParseResultToInflux(result, myExecuteProtecodeScanOptions.ProtecodeExcludeCVEs)
	fmt.Printf("Report Result: %v", m)

	//TODO write Report to file system that it can be used by the groovey step
	err := protecode.WriteVulnResultToFile(m, "VulnResult.txt", docFileWriter)
	if err != nil {
		panic(err)
	}

	//clean scan from server
	if myExecuteProtecodeScanOptions.CleanupMode == "complete" {
		fmt.Printf("Protecode scan successful. Deleting scan from server.")
		deleteScan(myExecuteProtecodeScanOptions, productId, client)
	}
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

func loadExistingProduct(config executeProtecodeScanOptions, client piperHttp.Client) string {
	var productId string = ""

	if config.ReuseExisting {

		response := loadExistingProductByFilename(config, client)
		// by definition we will take the first one and trigger rescan
		productId := response.Products[0].ProductId

		fmt.Printf("re-use existing Protecode scan - file: %v, group: %v, productId: %v", myExecuteProtecodeScanOptions.FilePath, myExecuteProtecodeScanOptions.ProtecodeGroup, productId)
	}

	return productId
}

func loadExistingProductByFilename(config executeProtecodeScanOptions, client piperHttp.Client) *protecode.ProteCodeProductData {

	protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/apps/", fmt.Sprintf("%v/", config.ProtecodeGroup), config.FilePath)
	headers := protecode.CreateRequestHeader(config.Verbose, protecode.GetBase64UserPassword(), map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	})

	r := protecode.SendApiRequest("GET", protecodeURL.String(), headers, client)

	return protecode.GetProteCodeProductData(*r)
}

func pullResult(config executeProtecodeScanOptions, productId string, client piperHttp.Client) protecode.ProteCodeResult {

	protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/", productId), "")
	headers := protecode.CreateRequestHeader(config.Verbose, protecode.GetBase64UserPassword(), map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	})

	r := protecode.SendApiRequest("GET", protecodeURL.String(), headers, client)

	return protecode.GetProteCodeResultData(*r).Result
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

func uploadScanFile(config executeProtecodeScanOptions) protecode.ProteCodeResult {

	cmdString := cmdStringUploadScanFile(config)
	response := protecode.CmdExecGetProtecodeResult("curl", cmdString)

	return response
}

func declareFetchUrl(config executeProtecodeScanOptions) protecode.ProteCodeResult {

	cmdString := cmdStringDeclareFetchUrl(config)
	response := protecode.CmdExecGetProtecodeResult("curl", cmdString)

	return response
}

func uploadScanOrDeclareFetch(config executeProtecodeScanOptions, productId string) string {

	// check if no existing is found or reuse existing is false
	if len(productId) == 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			result := declareFetchUrl(config)
			productId = result.ProductId
		} else {
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			result := uploadScanFile(config)
			productId = result.ProductId
		}
	}

	return productId
}

func pollForResult(config executeProtecodeScanOptions, productId string, client piperHttp.Client, dur time.Duration) protecode.ProteCodeResult {

	var response protecode.ProteCodeResult = protecode.ProteCodeResult{}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	ticks := dur / 10

	for i := ticks; i > 0; i-- {
		select {
		case t := <-ticker.C:
			fmt.Printf("Ticker %v", t)
			response = pullResult(config, productId, client)
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
		response = pullResult(config, productId, client)
		if len(response.Components) == 0 || response.Status == "B" {
			log.Entry().Fatal("No result for protecode scan")
			os.Exit(1)
		}
	}

	return response
}

func loadReport(config executeProtecodeScanOptions, productId string, client piperHttp.Client) *io.ReadCloser {

	protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/pdf-report", productId), "")
	headers := protecode.CreateRequestHeader(config.Verbose, protecode.GetBase64UserPassword(), map[string][]string{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Pragma":        []string{"no-cache"},
		"Outputfile":    []string{config.ReportFileName},
	})

	r := protecode.SendApiRequest("GET", protecodeURL.String(), headers, client)

	return r
}

func writeHttpResponseToFile(resp io.ReadCloser) {
	f, err := os.Create("protecode_output.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = io.Copy(f, resp)
	if err != nil {
		panic(err)
	}
}

func deleteScan(config executeProtecodeScanOptions, productId string, client piperHttp.Client) {

	switch config.CleanupMode {
	case "none":
	case "binary":
		return
	case "complete":
		protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/", productId), "")
		headers := protecode.CreateRequestHeader(config.Verbose, protecode.GetBase64UserPassword(), map[string][]string{})

		protecode.SendApiRequest("DELETE", protecodeURL.String(), headers, client)
		break
	default:
		log.Entry().Fatalf("Unknown cleanup mode %v", config.CleanupMode)
	}
}

func docFileWriter(filename string, b []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, b, perm)
}
