package protecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	statusBusy   = "B"
	statusReady  = "R"
	statusFailed = "F"

	endpointApps      = "/api/apps/%s/"
	endpointProduct   = "/api/product/%v/"
	endpointPdfReport = "/api/product/%v/pdf-report"
	endpointUpload    = "/api/upload/%v"
	endpointFetch     = "/api/fetch/"
)

func (pc *Protecode) send(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {
	r, err := pc.client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		return nil, err
	}
	return &r.Body, nil
}

func (pc *Protecode) pullResult(productID int) (ResultData, error) {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointProduct, productID), "", "")
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}
	r, err := pc.send(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return *new(ResultData), err
	}
	result := new(ResultData)
	pc.mapResponse(*r, result)

	return *result, nil

}

func (pc *Protecode) loadProductData(group string) *ProductData {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointApps, group), "", "")
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}

	r, err := pc.send(http.MethodGet, protecodeURL, headers)
	if err != nil {
		//TODO: return error
		pc.logger.WithError(err).Fatalf("Error during load existing product: %v", protecodeURL)
	}
	result := new(ProductData)
	pc.mapResponse(*r, result)

	return result
}

// DeleteScan deletes if configured the scan on the protecode server
func (pc *Protecode) DeleteScan(cleanupMode string, productID int) {
	//TODO: extract cleanupMode to step logic
	switch cleanupMode {
	case "none":
	case "binary":
	case "complete":
		pc.logger.Info("Deleting scan from server.")
		protecodeURL := pc.createURL(fmt.Sprintf(endpointProduct, productID), "", "")
		headers := map[string][]string{}
		//TODO: handle error
		pc.send(http.MethodDelete, protecodeURL, headers)
	default:
		pc.logger.Fatalf("Unknown cleanup mode %v", cleanupMode)
	}
}

// LoadReport loads the report of the protecode scan
func (pc *Protecode) LoadReport(reportFileName string, productID int) *io.ReadCloser {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointPdfReport, productID), "", "")
	headers := map[string][]string{
		"Cache-Control": {"no-cache, no-store, must-revalidate"},
		"Pragma":        {"no-cache"},
		"Outputfile":    {reportFileName},
	}

	readCloser, err := pc.send(http.MethodGet, protecodeURL, headers)
	if err != nil {
		//TODO: handle error
		pc.logger.WithError(err).Fatalf("It is not possible to load report %v", protecodeURL)
	}

	return readCloser
}

// UploadScanFile upload the scan file to the protecode server
func (pc *Protecode) UploadScanFile(cleanupMode, group, filePath, fileName string) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	protecodeURL := pc.createURL(fmt.Sprintf(endpointUpload, fileName), "", "")
	headers := map[string][]string{
		"Group":         {group},
		"Delete-Binary": {fmt.Sprintf("%v", deleteBinary)},
	}

	r, err := pc.client.UploadRequest(http.MethodPut, protecodeURL, filePath, "file", headers, nil)
	if err != nil {
		//TODO: handle error
		pc.logger.WithError(err).Fatalf("Error during %v upload request", protecodeURL)
	}
	pc.logger.Info("Upload successful")
	result := new(ResultData)
	pc.mapResponse(r.Body, result)

	return result
}

// DeclareFetchURL configures the fetch url for the protecode scan
func (pc *Protecode) DeclareFetchURL(cleanupMode, group, fetchURL string) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	protecodeURL := pc.createURL(endpointFetch, "", "")
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Group":         {group},
		"Delete-Binary": {fmt.Sprintf("%v", deleteBinary)},
		"Url":           {fetchURL},
	}

	r, err := pc.send(http.MethodPost, protecodeURL, headers)
	if err != nil {
		//TODO: handle error
		pc.logger.WithError(err).Fatalf("Error during declare fetch url: %v", protecodeURL)
	}

	result := new(ResultData)
	pc.mapResponse(*r, result)

	return result
}

func (pc *Protecode) mapResponse(r io.ReadCloser, response interface{}) {
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()
	if len(newStr) > 0 {

		unquoted, err := strconv.Unquote(newStr)
		if err != nil {
			err = json.Unmarshal([]byte(newStr), response)
			if err != nil {
				//TODO: return error
				pc.logger.WithError(err).Fatalf("Error during unqote response: %v", newStr)
			}
		} else {
			//TODO: return error
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			//TODO: return error
			pc.logger.WithError(err).Fatalf("Error during decode response: %v", newStr)
		}
	}
}
