package main

import (
	"bufio"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

const fileName = "integration_test_time_report.xlsx"

func main() {
	f := excelize.NewFile()
	defer f.SaveAs(fileName)
	logFile, _ := os.Open("integration/test.log")
	defer logFile.Close()
	scanner := bufio.NewScanner(logFile)
	iterator := 0
	for scanner.Scan() {
		if str := scanner.Text(); strings.Contains(str, "[testTimer]") {
			iterator++
			line := strings.Split(scanner.Text(), " ")
			testName := line[4]
			duration := line[7]
			for i := iterator; ; i++ {
				cell := fmt.Sprintf("A%d", i)
				v, _ := f.GetCellValue("Sheet1", cell)
				if len(v) == 0 {
					f.SetCellValue("Sheet1", cell, testName)
					durationFloat, _ := strconv.ParseFloat(duration, 64)
					f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i), durationFloat)
					break
				}
			}
		}
	}
	log.Entry().Infof("total %d", iterator)
}

func testTimer(testName string, start time.Time) {
	d := time.Now().Sub(start).Seconds()
	log.Entry().Infof(" [testTimer] %s completed in %v seconds", testName, d)
}

func timeNow() time.Time {
	return time.Now()
}
