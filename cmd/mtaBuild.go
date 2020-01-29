package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"os"
	"strings"
	"sync"
)

func mtaBuild(config mtaBuildOptions, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment) error {
	log.Entry().Info("Launching mta build")
	return runMtaBuild(config, commonPipelineEnvironment, &command.Command{})
}

func runMtaBuild(config mtaBuildOptions, commonPipelineEnvironment *mtaBuildCommonPipelineEnvironment,
	s shellRunner) error {

	prOut, pwOut := io.Pipe()
	prErr, pwErr := io.Pipe()

	s.Stdout(pwOut)
	s.Stderr(pwErr)

	var e, o string

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		buf := new(bytes.Buffer)
		r := io.TeeReader(prOut, os.Stderr)
		io.Copy(buf, r)
		o = buf.String()
		wg.Done()
	}()

	go func() {
		buf := new(bytes.Buffer)
		r := io.TeeReader(prErr, os.Stderr)
		io.Copy(buf, r)
		e = buf.String()
		wg.Done()
	}()

	//
	//mtaBuildTool := "classic"
	mtaBuildTool := "cloudMbt"
	buildTarget := "buildTarget"
	extensions := "ext"
	platform := "platform"
	//

	var mtaJar = "mta.jar"
	var mtaCall = `Echo "Hello MTA"`
	var options = []string{}

	if len(extensions) != 0 {
		options = append(options, fmt.Sprintf("--extension=%s", extensions))
	}

	switch mtaBuildTool {
	case "classic":
		options = append(options, fmt.Sprintf("--build-target=%s", buildTarget))
		mtaCall = fmt.Sprintf("java -jar %s %s build", mtaJar, strings.Join(options, " "))
	case "cloudMbt":
		options = append(options, fmt.Sprintf("--platform %s", platform))
		options = append(options, "--target ./")
		mtaCall = fmt.Sprintf("mbt build %s", strings.Join(options, " "))
	default:
		return fmt.Errorf("Unknown mta build tool: \"${%s}\"", mtaBuildTool)
	}

	log.Entry().Infof("Executing mta build call: \"%s\"", mtaCall)

	script := fmt.Sprintf(`#!/bin/bash
	export PATH=./node_modules/.bin:$PATH
	echo "[DEBUG] PATH: ${PATH}"
	%s`, mtaCall)

	if e := s.RunShell("/bin/bash", script); e != nil {
		return e
	}

	pwOut.Close()
	pwErr.Close()

	wg.Wait()

	mtarFilePath := "dummy.mtar"
	commonPipelineEnvironment.mtarFilePath = mtarFilePath
	return nil
}
