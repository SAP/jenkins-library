package cmd

import (
	//"encoding/base64"
	"fmt"
	"os"

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

func runProtecodeScan(myExecuteProtecodeScanOptions executeProtecodeScanOptions, command execRunner) {
	var response *protecode.ProteCodeResultData = &protecode.ProteCodeResultData{}
	var productId string = ""

	client := protecode.Client{}

	if myExecuteProtecodeScanOptions.ReuseExisting {

		response = loadExistingProductByFilename(myExecuteProtecodeScanOptions, client)
		// by definition we will take the first one and trigger rescan
		productId := response.Result.ProductId

		fmt.Printf("re-use existing Protecode scan - file: %v, group: %v, productId: %v", myExecuteProtecodeScanOptions.FilePath, myExecuteProtecodeScanOptions.ProtecodeGroup, productId)
	}
	//TODO check if the Result is available
	if !myExecuteProtecodeScanOptions.ReuseExisting {
		// Protecode scan
		if len(myExecuteProtecodeScanOptions.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", myExecuteProtecodeScanOptions.FetchURL, myExecuteProtecodeScanOptions.ProtecodeGroup)
			result := declareFetchUrl(myExecuteProtecodeScanOptions)
			productId = result.ProductId
		} else {
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", myExecuteProtecodeScanOptions.FilePath, myExecuteProtecodeScanOptions.ProtecodeGroup)
			result := uploadScanFile(myExecuteProtecodeScanOptions)
			productId = result.ProductId
		}
	}

	//pollForResult
	result := pollForResult(myExecuteProtecodeScanOptions, productId, client)

	if len(result.Status) > 0 || result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
		os.Exit(1)
		//Notify.error(this, "Protecode scan failed, please check the log and protecode backend for more details.")
	}

	//loadReport
	loadReport(myExecuteProtecodeScanOptions, productId, client)

	//count vulnerabilities
	m := protecode.ParseResultToInflux(result, myExecuteProtecodeScanOptions.ProtecodeExcludeCVEs)
	fmt.Printf("Report Result: %v", m)
	//TODO write Report to file system that it can be used by the groovey step

	if myExecuteProtecodeScanOptions.CleanupMode == "complete" {
		fmt.Printf("Protecode scan successful. Deleting scan from server.")
		deleteScan(myExecuteProtecodeScanOptions, productId, client)
	}
}

func loadExistingProductByFilename(config executeProtecodeScanOptions, client protecode.Client) *protecode.ProteCodeResultData {

	protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/apps/", fmt.Sprintf("%v/", config.ProtecodeGroup), config.FilePath)

	//TODO user and pwd from env cumulus
	headers := protecode.CreateRequestHeader(config.ProtecodeCredentialsID, config.Verbose, map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	})

	r := protecode.SendApiRequest("GET", protecodeURL.String(), headers, client)

	return protecode.GetProteCodeResultData(r)
}

func pullResult(config executeProtecodeScanOptions, productId string, client protecode.Client) protecode.ProteCodeResult {

	protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/", productId), "")
	//user and pwd from env
	headers := protecode.CreateRequestHeader(config.ProtecodeCredentialsID, config.Verbose, map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	})

	r := protecode.SendApiRequest("GET", protecodeURL.String(), headers, client)

	return protecode.GetProteCodeResultData(r).Result
}

//TODO
func getBase64UserPassword(config executeProtecodeScanOptions) string {
	if len(config.FilePath) > 0 {
		return "auth"
	}

	return "auth"
}

func cmdStringUploadScanFile(config executeProtecodeScanOptions) string {
	deleteBinary := (config.CleanupMode == "binary" || config.CleanupMode == "complete")

	cmdString := fmt.Sprintf("curl --insecure -H 'Authorization: Basic %v' %v -H 'Group: %v' -H 'Delete-Binary: %v' -T %v %v/api/upload/ --write-out '%vstatus=%v'",
		getBase64UserPassword(config), "" /* CallbackParameter */, config.ProtecodeGroup, deleteBinary, config.FilePath, config.ProtecodeServerURL, protecode.DELIMITER, "%{http_code}")

	//TODO PUT	
	return cmdString
}

func cmdStringDeclareFetchUrl(config executeProtecodeScanOptions) string {
	deleteBinary := (config.CleanupMode == "binary" || config.CleanupMode == "complete")
	cmdString := fmt.Sprintf("curl -X POST -H 'Authorization: Basic %v' %v -H 'Group: %v' -H 'Delete-Binary: %v' -H 'Url:%v'  %v/api/fetch/ --write-out '%vstatus=%v'",
		getBase64UserPassword(config), "" /* CallbackParameter */, config.ProtecodeGroup, deleteBinary, config.FetchURL, config.ProtecodeServerURL, protecode.DELIMITER, "%{http_code}")

	return cmdString
}

func uploadScanFile(config executeProtecodeScanOptions) protecode.ProteCodeResult {

	cmdString := cmdStringUploadScanFile(config)
	//TODO rework to sendApiRequest PUT
	response := protecode.CmdExecGetProtecodeResult("curl", cmdString)
	
	return response
}

func declareFetchUrl(config executeProtecodeScanOptions) protecode.ProteCodeResult {
	
	cmdString := cmdStringDeclareFetchUrl(config)
	//TODO rework to sendApiRequest POST

	sendApiRequedst "post" Authorization 
	response := protecode.CmdExecGetProtecodeResult("curl", cmdString)

	return response
}

func deleteScan(config executeProtecodeScanOptions, productId string, client protecode.Client) {

	switch config.CleanupMode {
	case "none":
	case "binary":
		return
	case "complete":
		protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/", productId), "")
		headers := protecode.CreateRequestHeader(config.ProtecodeCredentialsID, config.Verbose, map[string][]string{
		})

		protecode.SendApiRequest("DELETE", protecodeURL.String(), headers, client)
		break
	default:
		log.Entry().Fatalf("Unknown cleanup mode %v", config.CleanupMode)
	}
}

func pollForResult(config executeProtecodeScanOptions, productId string, client protecode.Client) protecode.ProteCodeResult {

	var response protecode.ProteCodeResult = protecode.ProteCodeResult{
		Status: "200",
	}

	return response
	//busy := true
	//try {
	//script.timeout(config.protecodeTimeoutMinutes.toInteger()) {
	//	script.echo "Polling result for Protecode scan - productId: ${productId}"
	//	def times = 0
	//	while (busy) {
	//		//def sleepSeconds = 10 * (2**times)
	//		script.sleep(60)
	//		json = pullResult(config, productId)
	//		busy = json?.results?.status == 'B'
	//		if (this.verbose)
	//			script.echo "Processing status for productId ${productId} is '${json?.results?.status}'"
	//		times++
	//		script.echo "After ${times} attempts artifact is still being processed"
	//	}
	//}
	//} catch (FlowInterruptedException e) {
	//	json = pullResult(config, productId)
	//	busy = json?.results?.status == 'B'
	//	if (busy)
	//		throw e
	//}
	//return (json != null ? json : [:])
}

func loadReport(config executeProtecodeScanOptions, productId string, client protecode.Client) {

	protecodeURL := protecode.CreateUrl(config.ProtecodeServerURL, "/api/product/", fmt.Sprintf("%v/pdf-report", productId), "")
	headers := protecode.CreateRequestHeader(config.ProtecodeCredentialsID, config.Verbose, map[string][]string{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Pragma":        []string{"no-cache"},
		"Outputfile":    []string{config.ReportFileName},
	})

	 protecode.SendApiRequest("GET", protecodeURL.String(), headers, client)
	 //TODO save file to filesystem
}
