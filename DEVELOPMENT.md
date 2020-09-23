# Development

**Table of contents:**

1. [Getting started](#getting-started)
1. [Build the project](#build-the-project)
1. [Generating step framework](#generating-step-framework)
1. [Best practices for writing piper-go steps](#best-practices-for-writing-piper-go-steps)
1. [Testing](#testing)
1. [Debugging](#debugging)
1. [Release](#release)
1. [Pipeline Configuration](#pipeline-configuration)

## Getting started

1. [Ramp up your development environment](#ramp-up)
1. [Get familiar with Go language](#go-basics)
1. Create [a GitHub account](https://github.com/join)
1. Setup [GitHub access via SSH](https://help.github.com/articles/connecting-to-github-with-ssh/)
1. [Create and checkout a repo fork](#checkout-your-fork)
1. Optional: [Get Jenkins related environment](#jenkins-environment)
1. Optional: [Get familiar with Jenkins Pipelines as Code](#jenkins-pipelines)

### Ramp up

First you need to set up an appropriate development environment:

1. Install Go, see [GO Getting Started](https://golang.org/doc/install)
1. Install an IDE with Go plugins, see for example [Go in Visual Studio Code](https://code.visualstudio.com/docs/languages/go)

### Go basics

In order to get yourself started, there is a lot of useful information out there.

As a first step to take we highly recommend the [Golang documentation](https://golang.org/doc/), especially [A Tour of Go](https://tour.golang.org/welcome/1).

We have a strong focus on high quality software and contributions without adequate tests will not be accepted.
There is an excellent resource which teaches Go using a test-driven approach: [Learn Go with Tests](https://github.com/quii/learn-go-with-tests)

### Checkout your fork

The project uses [Go modules](https://blog.golang.org/using-go-modules). Thus please make sure to **NOT** checkout the project into your [`GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH).

To check out this repository:

1. Create your own
    [fork of this repo](https://help.github.com/articles/fork-a-repo/)
1. Clone it to your machine, for example like:

```shell
mkdir -p ${HOME}/projects/jenkins-library
cd ${HOME}/projects
git clone git@github.com:${YOUR_GITHUB_USERNAME}/jenkins-library.git
cd jenkins-library
git remote add upstream git@github.com:sap/jenkins-library.git
git remote set-url --push upstream no_push
```

### Jenkins environment

If you want to contribute also to the Jenkins-specific parts like

* Jenkins library step
* Jenkins pipeline integration

you need to do the following in addition:

* [Install Groovy](https://groovy-lang.org/install.html)
* [Install Maven](https://maven.apache.org/install.html)
* Get a local Jenkins installed: Use for example [cx-server](https://github.com/SAP/devops-docker-cx-server)

### Jenkins pipelines

The Jenkins related parts depend on

* [Jenkins Pipelines as Code](https://jenkins.io/doc/book/pipeline-as-code/)
* [Jenkins Shared Libraries](https://jenkins.io/doc/book/pipeline/shared-libraries/)

You should get familiar with these concepts for contributing to the Jenkins-specific parts.

## Build the project

### Build the executable suitable for the CI/CD Linux target environments

Use Docker:

`docker build -t piper:latest .`

You can extract the binary using Docker means to your local filesystem:

```sh
docker create --name piper piper:latest
docker cp piper:/build/piper .
docker rm piper
```

## Generating step framework

The steps are generated based on the yaml files in `resources/metadata/` with the following command from the root of the project:

```bash
go generate
```

The yaml format is kept pretty close to Tekton's [task format](https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md).
Where the Tekton format was not sufficient some extenstions have been made.

Examples are:

* matadata - longDescription
* spec - inputs - secrets
* spec - containers
* spec - sidecars

There are certain extensions:

* **aliases** allow alternative parameter names also supporting deeper configuration structures. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/kubernetesdeploy.yaml)
* **resources** allow to read for example from a shared `commonPipelineEnvironment` which contains information which has been provided by a previous step in the pipeline via an output. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/githubrelease.yaml)
* **secrets** allow to specify references to Jenkins credentials which can be used in the `groovy` library. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/kubernetesdeploy.yaml)
* **outputs** allow to write to dedicated outputs like

  * Influx metrics. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/checkmarx.yaml)
  * Sharing data via `commonPipelineEnvironment` which can be used by another step as input

* **conditions** allow for example to specify in which case a certain container is used (depending on a configuration parameter). [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/kubernetesdeploy.yaml)

## Best practices for writing piper-go steps

1. [Logging](#logging)
1. [Error handling](#error-handling)

Implementing a new step starts by adding a new yaml file in `resources/metadata/` and running
the [step generator](#generating-step-framework). This creates most of the boiler-plate code for the
step's implementation in `cmd/`. There are four files per step based on the name given within the yaml:

1. `cmd/<step>.go` - contains the skeleton of your step implementation.
1. `cmd/<step>_test.go` - write your unit tests here.
1. `cmd/<step>_generated.go` - contains the generated boiler plate code, and a dedicated type definition for your step's options.
1. `cmd/<step>_generated_test.go` - contains a simple unit test for the generated part.

You never edit in the generated parts. If you need to make changes, you make them in the yaml and re-run the step
generator (which will of course not overwrite your implementation).

The file `cmd/<step>.go` initially contains two functions:

```golang
func step(options stepOptions, telemetryData *telemetry.CustomData) {
    err := runStep(&options, telemetryData)
    if err != nil {
        log.Entry().WithError(err).Fatal("step execution failed")
    }
}

func runStep(options *stepOptions, telemetryData *telemetry.CustomData) error {
}
```

The separation into these two functions facilitates unit tests and mocking. From your tests, you could call
`runStep()` with mocking instances of needed objects, while inside `step()`, you create runtime instances of these
objects.

### Logging

Logging is done via the [sirupsen/logrus](https://github.com/sirupsen/logrus) framework.
It can conveniently be accessed through:

```golang
import (
    "github.com/SAP/jenkins-library/pkg/log"
)

func myStep ...
    ...
    log.Entry().Info("This is my info.")
    ...
}
```

If a fatal error occurs your code should act similar to:

```golang
    ...
    if err != nil {
        log.Entry().
            WithError(err).
            Fatal("failed to execute step ...")
    }
```

Calling `Fatal` results in an `os.Exit(0)` and before exiting some cleanup actions (e.g. writing output data,
writing telemetry data if not deactivated by the user, ...) are performed.

### Error handling

In order to better understand the root cause of errors that occur, we wrap errors like

```golang
    f, err := os.Open(path)
    if err != nil {
        return errors.Wrapf(err, "open failed for %v", path)
    }
    defer f.Close()
```

We use [github.com/pkg/errors](https://github.com/pkg/errors) for that.

It has proven a good practice to bubble up errors until the runtime entry function  and only
there exit via the logging framework (see also [logging](#logging)).

### Error categories

For errors, we have a convenience function to set a pre-defined category once an error occurs:

```golang
log.SetErrorCategory(log.ErrorCompliance)
```

Error categories are defined in [`pkg/log/ErrorCategory`](pkg/log/errors.go).

With writing a fatal error

```golang
log.Entry().WithError(err).Fatal("the error message")
```
the category will be written into the file `errorDetails.json` and can be used from there in the further pipeline flow.
Writing the file is handled by [`pkg/log/FatalHook`](pkg/log/fatalHook.go).

## Testing

1. [Mocking](#mocking)
1. [Mockable Interface](#mockable-interface)
1. [Global function pointers](global-function-pointers)

Unit tests are done using basic `golang` means.

Additionally, we encourage you to use [github.com/stretchr/testify/assert](https://github.com/stretchr/testify/assert)
in order to have slimmer assertions if you like. A good pattern to follow is this:

```golang
func TestNameOfFunctionUnderTest(t *testing.T) {
    t.Run("A description of the test case", func(t *testing.T) {
        // init
        // test
        // assert
    })
    t.Run("Another test case", func(t *testing.T) {
        // init
        // test
        // assert
    })
}
```

This will also structure the test output for better readability.

### Mocking

Tests should be written only for the code of your step implementation, while any
external functionality should be mocked, in order to test all code paths including
the error cases.

There are (at least) two approaches for this:

#### Mockable Interface

In this approach you declare an interface that contains every external function
used within your step that you need to be able to mock. In addition, you declare a struct
which holds the data you need during runtime, and implement the interface with the "real"
functions. Here is an example to illustrate:

```golang
import (
    "github.com/SAP/jenkins-library/pkg/piperutils"
)

type myStepUtils interface {
    fileExists(path string) (bool, error)
    fileRead(path string) ([]byte, error)
}

type myUtilsData struct {
    fileUtils piperutils.Files
}

func (u *myUtilsData) fileExists(path string) (bool, error) {
    return u.fileUtils.FileExists(path)
}

func (u *myUtilsData) fileRead(path string) ([]byte, error) {
    return u.fileUtils.FileRead(path)
}
```

Then you create the runtime version of the utils data in your top-level entry function and
pass it to your `run*()` function:

```golang
func step(options stepOptions, _ *telemetry.CustomData) {
    utils := myUtilsData{
        fileUtils: piperutils.Files{},
    }
    err := runStep(&options, &utils)
    ...
}

func runStep(options *stepOptions, utils myStepUtils) error {
    ...
    exists, err := utils.fileExists(path)
    ...
}
```

In your tests, you would provide a mocking implementation of this interface and pass
instances of that to the functions under test. To better illustrate this, here is an example
for the interface above implemented in the `<step>_test.go` file:

```golang
type mockUtilsBundle struct {
    files map[string][]byte
}

func newMockUtilsBundle() mockUtilsBundle {
    utils := mockUtilsBundle{}
    utils.files = map[string][]byte{}
    return utils
}

func (m *mockUtilsBundle) fileExists(path string) (bool, error) {
    content := m.files[path]
    return content != nil, nil
}

func (m *mockUtilsBundle) fileRead(path string) ([]byte, error) {
    content := m.files[path]
    if content == nil {
        return nil, fmt.Errorf("could not read '%s': %w", path, os.ErrNotExist)
    }
    return content, nil
}

// This is how it would be used in tests:

func TestSomeFunction() {
    t.Run("Happy path", func(t *testing.T) {
        // init
        utils := newMockUtilsBundle()
        utils.files["some/path/file.xml"] = []byte(´content of the file´)
        // test
        err := someFunction(&utils)
        // assert
        assert.NoError(t, err)
    })
    t.Run("Error path", func(t *testing.T) {
        // init
        utils := newMockUtilsBundle()
        // test
        err := someFunction(&utils)
        // assert
        assert.EqualError(t, err, "could not read 'some/path/file.xml'")
    })
}
```

#### Global Function Pointers

An alternative approach are global function pointers:

```golang
import (
    FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

var fileUtilsExists = FileUtils.FileExists

func someFunction(options *stepOptions) error {
    ...
    exists, err := fileUtilsExists(path)
    ...
}
```

In your tests, you can then simply set the function pointer to a mocking implementation:

```golang
func TestSomeFunction() {
    t.Run("Happy path", func(t *testing.T) {
        // init
        originalFileExists := fileUtilsExists
        fileUtilsExists = func(filename string) (bool, error) {
            return true, nil
        }
        defer fileUtilsExists = originalFileExists
        // test
        err := someFunction(...)
        // assert
        assert.NoError(t, err)
    })
    t.Run("Error path", func(t *testing.T) {
        // init
        originalFileExists := fileUtilsExists
        fileUtilsExists = func(filename string) (bool, error) {
            return false, errors.New("something happened")
        }
        defer fileUtilsExists = originalFileExists
        // test
        err := someFunction(...)
        // assert
        assert.EqualError(t, err, "something happened")
    })
}
```

Both approaches have their own benefits. Global function pointers require less preparation
in the actual implementation and give great flexibility in the tests, while mocking interfaces
tend to result in more code re-use and slim down the tests. The mocking implementation of a
utils interface can facilitate implementations of related functions to be based on shared data.

## Debugging

Debugging can be initiated with VS code fairly easily. Compile the binary with specific compiler flags to turn off optimizations `go build -gcflags "all=-N -l" -o piper.exe`.

Modify the `launch.json` located in folder `.vscode` of your project root to point with `program` exactly to the binary that you just built with above command - must be an absolute path. Add any arguments required for the execution of the Piper step to `args`. What is separated with a blank on the command line must go into a separate string.

```javascript
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "exec",
            "program": "C:/CF@HCP/git/jenkins-library-public/piper.exe",
            "env": {},
            "args": ["checkmarxExecuteScan", "--password", "abcd", "--username", "1234", "--projectName", "testProject4711", "--serverUrl", "https://cx.server.com/"]
        }
    ]
}
```

Finally, set your breakpoints and use the `Launch` button in the VS code UI to start debugging.

## Release

Releases are performed using [Project "Piper" Action](https://github.com/SAP/project-piper-action).
We release on schedule (once a week) and on demand.
To perform a release, the respective action must be invoked for which a convenience script is available in `contrib/perform-release.sh`.
It requires a personal access token for GitHub with `repo` scope.
Example usage `PIPER_RELEASE_TOKEN=THIS_IS_MY_TOKEN contrib/perform-release.sh`.

## Pipeline Configuration

The pipeline configuration is organized in a hierarchical manner and configuration parameters are incorporated from multiple sources.
In general, there are four sources for configurations:

1. Directly passed step parameters
1. Project specific configuration placed in `.pipeline/config.yml`
1. Custom default configuration provided in `customDefaults` parameter of the project config or passed as parameter to the step `setupCommonPipelineEnvironment`
1. Default configuration from Piper library

For more information and examples on how to configure a project, please refer to the [configuration documentation](https://sap.github.io/jenkins-library/configuration/).

### Groovy vs. Go step configuration

The configuration of a project is, as of now, resolved separately for Groovy and Go steps.
There are, however, dependencies between the steps responsible for resolving the configuration.
The following provides an overview of the central components and their dependencies.

#### setupCommonPipelineEnvironment (Groovy)

The step `setupCommonPipelineEnvironment` initializes the `commonPipelineEnvironment` and `DefaultValueCache`.
Custom default configurations can be provided as parameters to `setupCommonPipelineEnvironment` or via the `customDefaults` parameter in project configuration.

#### DefaultValueCache (Groovy)

The `DefaultValueCache` caches the resolved (custom) default pipeline configuration and the list of configurations that contributed to the result.
On initialization, it merges the provided custom default configurations with the default configuration from Piper library, as per the hierarchical order.

Note, the list of configurations cached by `DefaultValueCache` is used to pass path to the (custom) default configurations of each Go step.
It only contains the paths of configurations which are **not** provided via `customDefaults` parameter of the project configuration, since the Go layer already resolves configurations provided via `customDefaults` parameter independently.

## Additional Developer Hints

You can find additional hints at [documentation/developer-hints](./documentation/developer_hints)
