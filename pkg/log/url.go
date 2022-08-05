package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mvdan.cc/xurls/v2"
	"os"
	"sync"
)

type (
	step struct {
		Step map[string]u `json:"step"`
	}
	u struct {
		URLs []string `json:"url"`
	}
)

const (
	urlLogFileName = "url-log.json"
)

type URLLogger struct {
	buf struct {
		data [][]byte
		sync.RWMutex
	}
	stepName string
}

func NewURLLogger(stepName string) *URLLogger {
	return &URLLogger{stepName: stepName}
}

func (cl *URLLogger) WriteURLsLogToJSON() error {
	cl.buf.Lock()
	defer cl.buf.Unlock()
	if len(cl.buf.data) == 0 {
		return nil
	}
	file, err := os.OpenFile(urlLogFileName, os.O_CREATE|os.O_RDWR, 0600)
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
	urlsLog := step{make(map[string]u)}
	if len(fileBuf) != 0 {
		err = json.Unmarshal(fileBuf, &urlsLog)
		if err != nil {
			return fmt.Errorf("can't unmarshal log: %w", err)
		}
		fileBuf = fileBuf[:0]
	}
	var urls []string
	if stepLogs, ok := urlsLog.Step[cl.stepName]; ok {
		urls = stepLogs.URLs
	}
	for _, url := range cl.buf.data {
		urls = append(urls, string(url))
	}
	urlsLog.Step[cl.stepName] = u{urls}
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
		return fmt.Errorf("failed to write log: %w", err)
	}
	return err
}

func (cl *URLLogger) Parse(buf bytes.Buffer) {
	cl.buf.Lock()
	defer cl.buf.Unlock()
	cl.buf.data = append(cl.buf.data, parseURLs(buf.Bytes())...)
}

func parseURLs(src []byte) [][]byte {
	return xurls.Strict().FindAll(src, -1)
}
