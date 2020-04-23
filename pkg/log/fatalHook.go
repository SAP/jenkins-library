package log

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type FatalHook struct {
	Path          string
	CorrelationID string
}

func (f *FatalHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.FatalLevel}
}

func (f *FatalHook) Fire(entry *logrus.Entry) error {
	details := entry.Data
	if details == nil {
		details = logrus.Fields{}
	}

	details["message"] = entry.Message
	details["result"] = "failure"
	details["correlationId"] = f.CorrelationID

	fileName := "errorDetails.json"
	if details["stepName"] != nil {
		fileName = fmt.Sprintf("%v_%v", fmt.Sprint(details["stepName"]), fileName)
	}
	filePath := filepath.Join(f.Path, fileName)

	_, err := ioutil.ReadFile(filePath)

	if err != nil {
		// ignore errors, since we don't want to break the logging flow
		errDetails, _ := json.Marshal(&details)
		ioutil.WriteFile(filePath, errDetails, 0655)
	}

	return nil
}
