# Development

**Table of contents:**

1. [Getting started](#getting-started)
1. [Build the project](#build-the-project_)
1. [Generating step framework](#generating-step-framework)
1. [Logging](#logging)
1. [Error handling](#error-handling)
1. [Debugging](#debugging)

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

Install Go, see [GO Getting Started](https://golang.org/doc/install)

Install an IDE with Go plugins, see for example [Go in Visual Studio Code](https://code.visualstudio.com/docs/languages/go)

### Go basics

In order to get yourself started, there is a lot of useful information out there.

As a first step to take we highly recommend the [Golang documentation](https://golang.org/doc/) especially, [A Tour of Go](https://tour.golang.org/welcome/1)

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
docker cp piper:/piper .
docker rm piper
```

## Generating step framework

The steps are generated based on the yaml files in `resources/metadata/` with the following command
`go run pkg/generator/step-metadata.go`.

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

## Logging

Logging is done through [sirupsen/logrus](https://github.com/sirupsen/logrus) framework.
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

Calling `Fatal` results in an `os.Exit(0)` and before exiting some cleanup actions (e.g. writing output data, writing telemetry data if not deactivated by the user, ...) are performed.

## Error handling

In order to better understand the root cause of errors that occur we wrap errors like

```golang
    f, err := os.Open(path)
    if err != nil {
        return errors.Wrapf(err, "open failed for %v", path)
    }
    defer f.Close()
```

We use [github.com/pkg/errors](https://github.com/pkg/errors) for that.

## Testing

Unit tests are done using basic `golang` means.

Additionally we encourage you to use [github.com/stretchr/testify/assert](https://github.com/stretchr/testify/assert) in order to have slimmer assertions if you like.

## Debugging

Debugging can be initiated with VS code fairly easily. Compile the binary with spcific compiler flags to turn off optimizations `go build -gcflags "all=-N -l" -o piper.exe`.

Modify the `launch.json` located in folder `.vscode` of your project root to point with `program` exatly to the binary that you just built with above command - must be an absolute path. In addition add any arguments required for the execution of the Piper step to `args`. What is separated with a blank on the command line must go into a separate string.

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

Finally set your breakpoints and use the `Launch` button in the VS code UI to start debugging.
