package codeql

import (
	"encoding/json"
	"io"
	"net/http"
)

type CodeqlSarifUploader interface {
	GetSarifStatus() (SarifFileInfo, error)
}

func NewCodeqlSarifUploaderInstance(url, token string) CodeqlSarifUploaderInstance {
	return CodeqlSarifUploaderInstance{
		url:   url,
		token: token,
	}
}

type CodeqlSarifUploaderInstance struct {
	url   string
	token string
}

func (codeqlSarifUploader *CodeqlSarifUploaderInstance) GetSarifStatus() (SarifFileInfo, error) {
	return getSarifUploadingStatus(codeqlSarifUploader.url, codeqlSarifUploader.token)
}

type SarifFileInfo struct {
	ProcessingStatus string   `json:"processing_status"`
	Errors           []string `json:"errors"`
}

const internalServerError = "Internal server error"

func getSarifUploadingStatus(sarifURL, token string) (SarifFileInfo, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", sarifURL, nil)
	if err != nil {
		return SarifFileInfo{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return SarifFileInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return SarifFileInfo{ProcessingStatus: internalServerError}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SarifFileInfo{}, err
	}

	sarifInfo := SarifFileInfo{}
	err = json.Unmarshal(body, &sarifInfo)
	if err != nil {
		return SarifFileInfo{}, err
	}
	return sarifInfo, nil
}
