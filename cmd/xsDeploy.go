package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

// DeployMode ...
type DeployMode int

const (
	// NoDeploy ...
	NoDeploy DeployMode = iota
	//Deploy ...
	Deploy DeployMode = iota
	//BGDeploy ...
	BGDeploy DeployMode = iota
)

// ValueOfMode ...
func ValueOfMode(str string) (DeployMode, error) {
	switch str {
	case "NONE":
		return NoDeploy, nil
	case "DEPLOY":
		return Deploy, nil
	case "BG_DEPLOY":
		return BGDeploy, nil
	default:
		return NoDeploy, errors.New(fmt.Sprintf("Unknown DeployMode: '%s'", str))
	}
}

// String ...
func (m DeployMode) String() string {
	return [...]string{
		"NONE",
		"DEPLOY",
		"BG_DEPLOY",
	}[m]
}

// Action ...
type Action int

const (
	//None ...
	None Action = iota
	//Resume ...
	Resume Action = iota
	//Abort ...
	Abort Action = iota
	//Retry ...
	Retry Action = iota
)

// ValueOfAction ...
func ValueOfAction(str string) (Action, error) {
	switch str {
	case "NONE":
		return None, nil
	case "RESUME":
		return Resume, nil
	case "ABORT":
		return Abort, nil
	case "RETRY":
		return Retry, nil

	default:
		return None, errors.New(fmt.Sprintf("Unknown Action: '%s'", str))
	}
}

// String ...
func (a Action) String() string {
	return [...]string{
		"NONE",
		"RESUME",
		"ABORT",
		"RETRY",
	}[a]
}

const loginScript = `#!/bin/bash
xs login -a {{.APIURL}} -u {{.Username}} -p '{{.Password}}' -o {{.Org}} -s {{.Space}} {{.LoginOpts}}
`

const logoutScript = `#!/bin/bash
xs logout`

const deployScript = `#!/bin/bash
xs {{.Mode}} {{.MtaPath}} {{.DeployOpts}}`

const completeScript = `#!/bin/bash
xs {{.Mode.GetDeployCommand}} -i {{.OperationID}} -a {{.Action.GetAction}}
`

func xsDeploy(config xsDeployOptions, telemetryData *telemetry.CustomData, piperEnvironment *xsDeployCommonPipelineEnvironment) {
	c := command.Command{}
	fileUtils := piperutils.Files{}
	err := runXsDeploy(config, piperEnvironment, &c, fileUtils, os.Remove, os.Stdout)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute xs deployment")
	}
}

func runXsDeploy(XsDeployOptions xsDeployOptions, piperEnvironment *xsDeployCommonPipelineEnvironment, s command.ShellRunner,
	fileUtils piperutils.FileUtils,
	fRemove func(string) error,
	stdout io.Writer) error {

	mode, err := ValueOfMode(XsDeployOptions.Mode)
	if err != nil {
		return fmt.Errorf("Extracting mode failed: '%s': %w", XsDeployOptions.Mode, err)
	}

	if mode == NoDeploy {
		log.Entry().Infof("Deployment skipped intentionally. Deploy mode '%s'", mode.String())
		return nil
	}

	action, err := ValueOfAction(XsDeployOptions.Action)
	if err != nil {
		return fmt.Errorf("Extracting action failed: '%s': %w", XsDeployOptions.Action, err)
	}

	if mode == Deploy && action != None {
		return errors.New(fmt.Sprintf("Cannot perform action '%s' in mode '%s'. Only action '%s' is allowed.", action, mode, None))
	}

	log.Entry().Debugf("Mode: '%s', Action: '%s'", mode, action)

	performLogin := mode == Deploy || (mode == BGDeploy && !(action == Resume || action == Abort))
	performLogout := mode == Deploy || (mode == BGDeploy && action != None)
	log.Entry().Debugf("performLogin: %t, performLogout: %t", performLogin, performLogout)

	{
		exists, e := fileUtils.FileExists(XsDeployOptions.MtaPath)
		if e != nil {
			return e
		}
		if action == None && !exists {
			return errors.New(fmt.Sprintf("Deployable '%s' does not exist", XsDeployOptions.MtaPath))
		}
	}

	if action != None && len(XsDeployOptions.OperationID) == 0 {
		return errors.New(fmt.Sprintf("OperationID was not provided. This is required for action '%s'.", action))
	}

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
		if _, err := io.Copy(buf, r); err != nil {
			log.Entry().Warningf("failed to copy buffer")
		}
		o = buf.String()
		wg.Done()
	}()

	go func() {
		buf := new(bytes.Buffer)
		r := io.TeeReader(prErr, os.Stderr)
		if _, err := io.Copy(buf, r); err != nil {
			log.Entry().Warningf("failed to copy buffer")
		}
		e = buf.String()
		wg.Done()
	}()

	var loginErr error

	xsSessionFile := ".xsconfig"
	if len(XsDeployOptions.XsSessionFile) > 0 {
		xsSessionFile = XsDeployOptions.XsSessionFile
	}

	if performLogin {
		loginErr = xsLogin(XsDeployOptions, s)
		if loginErr == nil {
			err = copyFileFromHomeToPwd(xsSessionFile, fileUtils)
		}
	}

	if loginErr == nil && err == nil {

		{
			exists, e := fileUtils.FileExists(xsSessionFile)
			if e != nil {
				return e
			}
			if !exists {
				return fmt.Errorf("xs session file does not exist (%s)", xsSessionFile)
			}
		}

		copyFileFromPwdToHome(xsSessionFile, fileUtils)

		switch action {
		case Resume, Abort, Retry:
			err = complete(mode, action, XsDeployOptions.OperationID, s)
		default:
			err = deploy(mode, XsDeployOptions, s)
		}
	}

	if loginErr == nil && (performLogout || err != nil) {
		if logoutErr := xsLogout(XsDeployOptions, s); logoutErr != nil {
			if err == nil {
				err = logoutErr
			}
		} else {

			// we delete the xs session file from workspace. From home directory it is deleted by the
			// xs command itself.
			if e := fRemove(xsSessionFile); e != nil {
				err = e
			}
			log.Entry().Debugf("xs session file '%s' has been deleted from workspace", xsSessionFile)
		}
	} else {
		if loginErr != nil {
			log.Entry().Info("Logout skipped since login did not succeed.")
		} else if !performLogout {
			log.Entry().Info("Logout skipped in order to be able to resume or abort later")
		}
	}

	if err == nil {
		err = loginErr
	}

	if err != nil {
		if e := handleLog(fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".xs_logs")); e != nil {
			log.Entry().Warningf("Cannot provide the logs: %s", e.Error())
		}
	}

	pwOut.Close()
	pwErr.Close()

	wg.Wait()

	if err == nil && (mode == BGDeploy && action == None) {
		piperEnvironment.operationID = retrieveOperationID(o, XsDeployOptions.OperationIDLogPattern)
		if len(piperEnvironment.operationID) == 0 && err == nil {
			err = errors.New("No operationID found")
		}
		XsDeployOptions.OperationID = piperEnvironment.operationID // for backward compatibility as long as we render that struc to stdout (printStatus)
	}

	if err != nil {
		log.Entry().Errorf("An error occurred. Stdout from underlying process: >>%s<<. Stderr from underlying process: >>%s<<", o, e)
	}

	if e := printStatus(XsDeployOptions, stdout); e != nil {
		if err == nil {
			err = e
		}
	}

	return err
}

func printStatus(XsDeployOptions xsDeployOptions, stdout io.Writer) error {
	XsDeployOptionsCopy := XsDeployOptions
	XsDeployOptionsCopy.Password = ""

	var e error
	if b, e := json.Marshal(XsDeployOptionsCopy); e == nil {
		fmt.Fprintln(stdout, string(b))
	}
	return e
}

func handleLog(logDir string) error {

	if _, e := os.Stat(logDir); !os.IsNotExist(e) {
		log.Entry().Warningf("Here are the logs (%s):", logDir)

		logFiles, e := os.ReadDir(logDir)

		if e != nil {
			return e
		}

		if len(logFiles) == 0 {
			log.Entry().Warningf("Cannot provide xs logs. No log files found inside '%s'.", logDir)
		}

		for _, logFile := range logFiles {
			buf := make([]byte, 32*1024)
			log.Entry().Infof("File: '%s'", logFile.Name())
			if f, e := os.Open(fmt.Sprintf("%s/%s", logDir, logFile.Name())); e == nil {
				for {
					if n, e := f.Read(buf); e != nil {
						if e == io.EOF {
							break
						}
						return e
					} else {
						os.Stderr.WriteString(string(buf[:n]))
					}
				}
			} else {
				return e
			}
		}
	} else {
		log.Entry().Warningf("Cannot provide xs logs. Log directory '%s' does not exist.", logDir)
	}
	return nil
}

func retrieveOperationID(deployLog, pattern string) string {
	re := regexp.MustCompile(pattern)
	lines := strings.Split(deployLog, "\n")
	var operationID string
	for _, line := range lines {
		matched := re.FindStringSubmatch(line)
		if len(matched) >= 1 {
			operationID = matched[1]
			break
		}
	}

	if len(operationID) > 0 {
		log.Entry().Infof("Operation identifier: '%s'", operationID)
	} else {
		log.Entry().Infof("No operation identifier found in >>>>%s<<<<.", deployLog)
	}

	return operationID
}

func xsLogin(XsDeployOptions xsDeployOptions, s command.ShellRunner) error {

	log.Entry().Debugf("Performing xs login. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	if e := executeCmd("login", loginScript, XsDeployOptions, s); e != nil {
		log.Entry().Errorf("xs login failed: %s", e.Error())
		return e
	}

	log.Entry().Infof("xs login has been performed. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	return nil
}

func xsLogout(XsDeployOptions xsDeployOptions, s command.ShellRunner) error {

	log.Entry().Debug("Performing xs logout.")

	if e := executeCmd("logout", logoutScript, XsDeployOptions, s); e != nil {
		return e
	}
	log.Entry().Info("xs logout has been performed")

	return nil
}

func deploy(mode DeployMode, XsDeployOptions xsDeployOptions, s command.ShellRunner) error {

	deployCommand, err := mode.GetDeployCommand()
	if err != nil {
		return err
	}

	type deployProperties struct {
		xsDeployOptions
		Mode string
	}

	log.Entry().Infof("Performing xs %s.", deployCommand)
	if e := executeCmd("deploy", deployScript, deployProperties{xsDeployOptions: XsDeployOptions, Mode: deployCommand}, s); e != nil {
		return e
	}
	log.Entry().Infof("xs %s performed.", deployCommand)

	return nil
}

func complete(mode DeployMode, action Action, operationID string, s command.ShellRunner) error {
	log.Entry().Debugf("Performing xs %s", action)

	type completeProperties struct {
		xsDeployOptions
		Mode        DeployMode
		Action      Action
		OperationID string
	}

	CompleteProperties := completeProperties{Mode: mode, Action: action, OperationID: operationID}

	if e := executeCmd("complete", completeScript, CompleteProperties, s); e != nil {
		return e
	}

	return nil
}

func executeCmd(templateID string, commandPattern string, properties interface{}, s command.ShellRunner) error {

	tmpl, e := template.New(templateID).Parse(commandPattern)
	if e != nil {
		return e
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, properties); err != nil {
		return err
	}
	if e := s.RunShell("/bin/bash", script.String()); e != nil {
		return e
	}

	return nil
}

func copyFileFromHomeToPwd(xsSessionFile string, fileUtils piperutils.FileUtils) error {
	src, dest := fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile), xsSessionFile
	log.Entry().Debugf("Copying xs session file from home directory ('%s') to workspace ('%s')", src, dest)
	if _, err := fileUtils.Copy(src, dest); err != nil {
		return fmt.Errorf("Cannot copy xssession file from home directory ('%s') to workspace ('%s'): %w", src, dest, err)
	}
	log.Entry().Debugf("xs session file copied from home directory ('%s') to workspace ('%s')", src, dest)
	return nil
}

func copyFileFromPwdToHome(xsSessionFile string, fileUtils piperutils.FileUtils) error {

	//
	// We rely on running inside a docker container which is discarded after a single use.
	// In general it is not a good idea to update files in the build users home directory in case
	// we are on an infrastructure which is used not only for single builds since updates at that level
	// affects also other builds.
	//

	src, dest := xsSessionFile, fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile)
	log.Entry().Debugf("Copying xs session file from workspace ('%s') to home directory ('%s')", src, dest)
	if _, err := fileUtils.Copy(src, dest); err != nil {
		return fmt.Errorf("Cannot copy xssession file from workspace ('%s') to home directory ('%s'): %w", src, dest, err)
	}
	log.Entry().Debugf("xs session file copied from workspace ('%s') to home directory ('%s')", src, dest)
	return nil
}

// GetAction ...
func (a Action) GetAction() (string, error) {
	switch a {
	case Resume, Abort, Retry:
		return strings.ToLower(a.String()), nil
	}
	return "", errors.New(fmt.Sprintf("Invalid deploy mode: '%s'.", a))

}

// GetDeployCommand ...
func (m DeployMode) GetDeployCommand() (string, error) {

	switch m {
	case Deploy:
		return "deploy", nil
	case BGDeploy:
		return "bg-deploy", nil
	}
	return "", errors.New(fmt.Sprintf("Invalid deploy mode: '%s'.", m))
}
