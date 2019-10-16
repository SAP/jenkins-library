# Development

**Table of contents:**

1. [Getting started](#getting-started)
1. [Build the project](#build-the-project_)
1. [Logging](#logging)
1. [Error handling](#error-handling)

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
* Get a local Jenkins installed: Use for example [cx-server](toDo: add link)

### Jenkins pipelines

The Jenkins related parts depend on 

* [Jenkins Pipelines as Code](https://jenkins.io/doc/book/pipeline-as-code/)
* [Jenkins Shared Libraries](https://jenkins.io/doc/book/pipeline/shared-libraries/)

You should get familiar with these concepts for contributing to the Jenkins-specific parts.

## Build the project

### Build the executable suitable for the CI/CD Linux target environments:

Use Docker:

`docker build -t piper:latest .`

You can extract the binary using Docker means to your local filesystem:

```
docker create --name piper piper:latest
docker cp piper:/piper .
docker rm piper
```

### Build the executable suitable for your local environment (e.g. Windows)

`go build -o piper.exe`

## Generating step framework

The steps are generated based on the yaml files in `resources/metadata/` with the following command
`go run pkg/generator/step-metadata.go`.

The yaml format is kept pretty close to Tekton's [task format](https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md).<br />
Where the Tekton format was not sufficient some extenstions have been made.<br />
Examples are:

* matadata - longDescription
* spec - inputs - secrets
* spec - containers
* spec - sidecars

## Logging

to be added

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