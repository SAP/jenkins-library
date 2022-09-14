package main

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"time"
)

func testTimer(testName string, start time.Time) {
	log.Entry().Infof("%s completed in %v", testName, time.Now().Sub(start).Minutes())
}
