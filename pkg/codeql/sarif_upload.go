package codeql

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	sarifUploadComplete = "complete"
	sarifUploadFailed   = "failed"
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

func WaitSarifUploaded(maxRetries, checkRetryInterval int, codeqlSarifUploader CodeqlSarifUploader) error {
	retryInterval := time.Duration(checkRetryInterval) * time.Second

	log.Entry().Info("waiting for the SARIF to upload")
	i := 1
	for {
		sarifStatus, err := codeqlSarifUploader.GetSarifStatus()
		if err != nil {
			return err
		}
		log.Entry().Infof("the SARIF processing status: %s", sarifStatus.ProcessingStatus)
		if sarifStatus.ProcessingStatus == sarifUploadComplete {
			return nil
		}
		if sarifStatus.ProcessingStatus == sarifUploadFailed {
			for e := range sarifStatus.Errors {
				log.Entry().Error(e)
			}
			return errors.New("failed to upload sarif file")
		}
		if i <= maxRetries {
			log.Entry().Infof("still waiting for the SARIF to upload: retrying in %d seconds... (retry %d/%d)", checkRetryInterval, i, maxRetries)
			time.Sleep(retryInterval)
			i++
			continue
		}
		return errors.New("failed to check sarif uploading status: max retries reached")
	}
}
