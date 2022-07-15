package cumuluslog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mvdan.cc/xurls/v2"
	"os"
)

type (
	URLsLog struct {
		Step map[string]Logs `json:"step"`
	}
	Logs struct {
		URLs []string `json:"urls"`
	}
)

const (
	urlsLogFileName = "urls-log.json"
)

func WriteURLsLogToJSON(urlsBuf [][]byte, stepName string) error {
	file, err := os.OpenFile(urlsLogFileName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		dErr := file.Close()
		if dErr != nil {
			err = fmt.Errorf("can't close file: %w", dErr)
		}
	}()
	fileBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can't read from gile: %w", err)
	}
	urlsLog := URLsLog{make(map[string]Logs)}
	if len(fileBuf) != 0 {
		err = json.Unmarshal(fileBuf, &urlsLog)
		if err != nil {
			return fmt.Errorf("can't unmarshal Logs: %w", err)
		}
		fileBuf = fileBuf[:0]
	}
	var urls []string
	if stepLogs, ok := urlsLog.Step[stepName]; ok {
		urls = stepLogs.URLs
	}
	for _, url := range urlsBuf {
		urls = append(urls, string(url))
	}
	urlsLog.Step[stepName] = Logs{urls}
	encoderBuf := bytes.NewBuffer(fileBuf)
	jsonEncoder := json.NewEncoder(encoderBuf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", " ")
	err = jsonEncoder.Encode(urlsLog)
	if err != nil {
		return fmt.Errorf("json encode error: %w", err)
	}
	_, err = file.WriteAt(encoderBuf.Bytes(), 0)
	if err != nil {
		return fmt.Errorf("failed to write Logs: %w", err)
	}
	return err
}

func ParseURLs(src []byte) [][]byte {
	return xurls.Strict().FindAll(src, -1)
}
