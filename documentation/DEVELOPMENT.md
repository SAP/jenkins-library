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
1. [Security Setup](#security-setup)
1. [Best practices for writing groovy](#best-practices-for-writing-groovy)

## Getting started

1. [Ramp up your development environment](#ramp-up)
1. [Get familiar with Go language](#go-basics)
1. Create [a GitHub account](https://github.com/join)
1. Setup [GitHub access via SSH](https://help.github.com/articles/connecting-to-github-with-ssh/)
1. [Create and checkout a repo fork](#checkout-your-fork)
1. [Editorconfig](#editorconfig)
1. [Commit message style](#commit-message-style)
1. Optional: [Get Jenkins related environment](#jenkins-environment)
1. Optional: [Get familiar with Jenkins Pipelines as Code](#jenkins-pipelines)

### Ramp up

First you need to set up an appropriate development environment:

1. Install Go, see [GO Getting Started](https://golang.org/doc/install)
1. Install an IDE with Go plugins, see for example [Go in Visual Studio Code](https://code.visualstudio.com/docs/languages/go)

**Note:** The Go version to be used is the one specified in the "go.mod" file.

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

### EditorConfig

To ensure a common file format, there is a `.editorConfig` file [in place](../.editorconfig). To respect this file, [check](http://editorconfig.org/#download) if your editor does support it natively or you need to download a plugin.

### Commit Message Style

Write [meaningful commit messages](http://who-t.blogspot.de/2009/12/on-commit-messages.html) and [adhere to standard formatting](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

Good commit messages speed up the review process and help to keep this project maintainable in the long term.

### Jenkins environment

If you want to contribute also to the Jenkins-specific parts like

* Jenkins library step
* Jenkins pipeline integration

you need to do the following in addition:

* [Install Groovy](https://groovy-lang.org/install.html)
* [Install Maven](https://maven.apache.org/install.html)
* Get a local Jenkins installed

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

* **aliases** allow alternative parameter names also supporting deeper configuration structures. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/kubernetesDeploy.yaml)
* **resources** allow to read for example from a shared `commonPipelineEnvironment` which contains information which has been provided by a previous step in the pipeline via an output. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/githubPublishRelease.yaml)
* **secrets** allow to specify references to Jenkins credentials which can be used in the `groovy` library. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/kubernetesDeploy.yaml)
* **outputs** allow to write to dedicated outputs like

  * Influx metrics. [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/checkmarxExecuteScan.yaml)
  * Sharing data via `commonPipelineEnvironment` which can be used by another step as input

* **conditions** allow for example to specify in which case a certain container is used (depending on a configuration parameter). [Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/kubernetesDeploy.yaml)

## Best practices for writing piper-go steps

1. [Logging](#logging)
1. [Error handling](#error-handling)
1. [HTTP calls](#http-calls)

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
        return fmt.Errorf("open failed for %v: %w", path, err)
    }
    defer f.Close()
```

We use standard library fmt.Error for that.

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

### HTTP calls

All HTTP(S) interactions with other systems should be leveraging the [`pkg/http`](pkg/http) to enable capabilities provided
centrally like automatic retries in case of intermittend HTTP errors or individual and optimized timout or logging capabilities.
The HTTP package provides a thin wrapper around the standard golang `net/http` package adding just the right bit of sugar on top to
have more control on common behaviors.

### Automatic retries

Automatic retries have been implemented based on [hashicorp's retryable HTTP client for golang](https://github.com/hashicorp/go-retryablehttp)
with some extensions and customizations to the HTTP status codes being retried as well as to improve some service specific error situations.
The client by default retries 15 times until it gives up and regards a specific communication event as being not recoverable. If you know by heart that
your service is much more stable and cloud live without retry handling or a specifically lower amout of retries, you can easily customize behavior via the
`ClientOptions` as shown in the sample below:

```golang
clientOptions := piperhttp.ClientOptions{}
clientOptions.MaxRetries = -1
httpClient.SetOptions(clientOptions)
```

## Testing

1. [Mocking](#mocking)
1. [Mockable Interface](#mockable-interface)
1. [Global function pointers](#global-function-pointers)
1. [Test Parallelization](#test-parallelization)

Unit tests are done using basic `golang` means. As the test files are tagged add the corresponding tag to the run command, as for example `go test -run ^TestRunAbapAddonAssemblyKitCheck$ github.com/SAP/jenkins-library/cmd -tags=unit`. In VSCode this can be done by adding the flag `"-tags=unit"` to the list of `"go.testFlags"` in the `settings.json` of the go extension.

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

### Test Parallelization

Tests that can be executed in parallel should be marked as such.
With the command `t.Parallel()` the test framework can be notified that this test can run in parallel, and it can start running the next test.
([Example in Stackoverflow](https://stackoverflow.com/questions/44325232/are-tests-executed-in-parallel-in-go-or-one-by-one))
Therefore, this command shall be called at the beginning of a test method **and also** in each `t.Run()` sub tests.
See also the [documentation](https://golang.org/pkg/testing/#T.Parallel) for `t.Parallel()` and `t.Run()`.

```go
func TestMethod(t *testing.T) {
    t.Parallel() // indicates that this method can run parallel to other methods

    t.Run("sub test 1", func(t *testing.T){
        t.Parallel() // indicates that this sub test can run parallel to other sub tests
        // execute test
    })

    t.Run("sub test 2", func(t *testing.T){
        t.Parallel() // indicates that this sub test can run parallel to other sub tests
        // execute test
    })
}
```

Go will first execute the non-parallelized tests in sequence and afterwards execute all the parallel tests in parallel, limited by the default number of parallel executions.

It is important that tests executed in parallel use the variable values actually meant to be visible to them.
Especially in table tests, it can happen easily that a variable injected into the `t.Run()`-closure via the outer scope is changed before or while the closure executes.
To prevent this, it is possible to create shadowing instances of variables in the body of the test loop.
(See [blog about it](https://eleni.blog/2019/05/11/parallel-test-execution-in-go/).)
At the minimum, you need to capture the test case value from the loop iteration variable, by shadowing this variable in the loop body.
Inside the `t.Run()` closure, this shadow copy is visible, and cannot be overwritten by later loop iterations.
If you do not make this shadowing copy, what is visible in the closure is the variable which gets re-assigned with a new value in each loop iteration.
The value of this variable is then not fixed for the test run.

```go
func TestMethod(t *testing.T) {
    t.Parallel() // indicates that this method can parallel to other methods
    testCases := []struct {
        Name string
    }{
        {
            Name: "Name1"
        },
        {
            Name: "Name2"
        },
    }

    for _, testCase := range testCases { // testCase defined here is re-assigned in each iteration
        testCase := testCase // define new variable within loop to detach from overwriting of the outer testCase variable by next loop iteration
        // The same variable name "testCase" is used for convenience.
        t.Run(testCase.Name, func(t *testing.T) {
            t.Parallel() // indicates that this sub test can run parallel to other sub tests
            // execute test
        })
    }
}
```

### Test pipeline for your fork (Jenkins)

Piper is ececuting the steps of each stage within a container. If you want to test your developments you have to ensure they are part of the image which is used in your test pipeline.

#### Testing Pipeline or Stage Definition changes (Jenkins)

As the pipeline and stage definitions (e.g. \*Pipeline\*Stage\*.groovy files in the vars folder) are directly executed you can easily test them just by referencing to your repo/branch/tag in the jenkinsfile.

```groovy
@Library('my-piper-lib-os-fork@MyTest') _

abapEnvironmentPipeline script: this
```

#### Testing changes on Step Level (Jenkins)

To trigger the creation of a "custom" container with your changes you can reuse a feature in piper which is originally meant for executing the integration tests. If the environment variables 'REPOSITORY_UNDER_TEST' (pointing to your forked repo) and 'LIBRARY_VERSION_UNDER_TEST' (pointing to a tag in your forked repo) are set a corresponding container gets created on the fly upon first usage in the pipeline. The drawback is that this takes extra time (1-2 minutes) you have to spend for every execution of the pipeline.

```groovy
@Library('piper-lib-os') _

env.REPOSITORY_UNDER_TEST       = 'myfork' // e.g. 'myUser/jenkins-library'
env.LIBRARY_VERSION_UNDER_TEST  = 'MyTag'

abapEnvironmentPipeline script: this
```

#### Using Parameterized Pipelines (Jenkins)

For test purpose it can be useful to utilize a parameterized pipeline. E.g. to toggle creation of the custom container:

```groovy
@Library('my-piper-lib-os-fork@MyTest') _

properties([
    parameters([
        booleanParam(name: 'toggleSomething', defaultValue: false, description: 'dito'),
        booleanParam(name: 'testPiperFork', defaultValue: false, description: 'dito'),
        string(name: 'repoUnderTest', defaultValue: '<MyUser>/jenkins-library', description: 'dito'),
        string(name: 'tag', defaultValue: 'MyTest', description: 'dito')
    ])
])

if (params.testPiperFork == true) {
    env.REPOSITORY_UNDER_TEST       = params.repoUnderTest
    env.LIBRARY_VERSION_UNDER_TEST  = params.tag
}

abapEnvironmentPipeline script: this
```

or skipping steps/stages with the help of extensions:

```groovy
void call(Map piperParams) {
  echo "Start - Extension for stage: ${piperParams.stageName}"

  if (params.toggleSomething == true) {
    // do something
    echo "now execute original stage as defined in the template"
    piperParams.originalStage()
  } else {
    // do something else
    // e.g. only this singele step of the stage
    somePiperStep( script: piperParams.script, someConfigParameter: '<...>' )
  }

  echo "End - Extension for stage: ${piperParams.stageName}"
}
return this
```

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

Releases are performed using GitHub workflows and [Project "Piper" Action](https://github.com/SAP/project-piper-action).

There are two different workflows:

- [weekly release workflow](.github/workflows/release-go.yml) running every Monday at 09:00 UTC.
- [commit-based workflow](.github/workflows/upload-go-master.yml) which releases the binary as `piper_master` on the [latest release](https://github.com/SAP/jenkins-library/releases/latest).

It is also possible to release on demand using the `contrib/perform-release.sh` script with a personal access token (`repo` scope).

```
PIPER_RELEASE_TOKEN=<token> contrib/perform-release.sh
```

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

## Security Setup

Here some hints and tricks are described to enhance the security within the development process.

1. [Signing Commits](#signing-commits)

### Signing Commits

In git, commits can be [signed](https://git-scm.com/book/en/v2/Git-Tools-Signing-Your-Work) to guarantee that that changes were made by the person named in the commit.
The name and email used for commits can be easily modified in the local git setup and afterwards it cannot be distinguished anymore if the commit was done by the real person or by some potential attacker.

In Windows, this can be done via [GnuPG](https://www.gnupg.org/(en)/download/index.html).
Download and install the tool.
Via the manager tool *Kleopatra* a new key pair can be easily created with a little wizard.
Make sure that the name and email are the ones used in your git.

The public key must then be added to the github's GPG section.
The private key should be kept in a backup as this signature is bound to you and not your machine.

The only thing left are some changes in the *.gitconfig* file.
The file shall be located in your user directory.
It might look something like the following.
All parts that are not relevant for signing were removed.

```
[user]
  name = My Name
  email = my.name@sap.com
  # Hash or email of you GPG key
  signingkey = D3CF72CC4006DE245C049566242831AEEE9DA2DD
[commit]
  # enable signing for commits
  gpgsign = true
[tag]
  # enable signing for tags (note the capital S)
  gpgSign = true
[gpg]
  # Windows was not able to find the private key. Setting the gpg command to use solved this.
  program = C:\\Program Files (x86)\\GnuPG\\bin\\gpg.exe
```

Add the three to four lines to you git config and this will do the necessary such that all your commits will be signed.

## Best practices for writing groovy

New steps should be written in go.

### Coding pattern

Pipeline steps must not make use of return values. The pattern for sharing parameters between pipeline steps or between a pipeline step and a pipeline script is sharing values via the [`commonPipelineEnvironment`](../vars/commonPipelineEnvironment.groovy). Since there is no return value from a pipeline step the return value of a pipeline step is already `void` rather than `def`.

### Jenkins credential handling

References to Jenkins credentials should have meaningful names.

We are using the following approach for naming Jenkins credentials:

For username/password credentials:
`<tool>CredentialsId` like e.g. `neoCredentialsId`

For other cases we add further information to the name like:

* `gitSshCredentialsId` for ssh credentials
* `githubTokenCredentialsId`for token/string credentials
* `gcpFileCredentialsId` for file credentials

### Code Style

Generally, the code should follow any stylistic and architectural guidelines prescribed by the project. In the absence of guidelines, mimic the styles and patterns in the existing code-base.

The intention of this section is to describe the code style for this project. As reference document, the [Groovy's style guide](http://groovy-lang.org/style-guide.html) was taken. For further reading about Groovy's syntax and examples, please refer to this guide.

This project is intended to run in Jenkins [[2]](https://jenkins.io/doc/book/getting-started/) as part of a Jenkins Pipeline [[3]](https://jenkins.io/doc/book/pipeline/). It is composed by Jenkins Pipeline's syntax, Groovy's syntax and Java's syntax.

Some Groovy's syntax is not yet supported by Jenkins. It is also the intention of this section to remark which Groovy's syntax is not yet supported by Jenkins.

As Groovy supports 99% of Java’s syntax [[1]](http://groovy-lang.org/style-guide.html), many Java developers tend to write Groovy code using Java's syntax. Such a developer should also consider the following code style for this project.

#### General remarks

Variables, methods, types and so on shall have meaningful self describing names. Doing so makes understanding code easier and requires less commenting. It helps people who did not write the code to understand it better.

Code shall contain comments to explain the intention of the code when it is unclear what the intention of the author was. In such cases, comments should describe the "why" and not the "what" (that is in the code already).

#### Omit semicolons

#### Use the return keyword

In Groovy it is optional to use the *return* keyword. Use explicitly the *return* keyword for better readability.

#### Use def

When using *def* in Groovy, the type is Object. Using *def* simplifies the code, for example imports are not needed, and therefore the development is faster.

#### Do not use a visibility modifier for public classes and methods

By default, classes and methods are public, the use of the public modifier is not needed.

#### Do not omit parentheses for Groovy methods

In Groovy is possible to omit parentheses for top-level expressions, but [Jenkins Pipeline's syntax](https://jenkins.io/doc/book/pipeline/syntax/) use a block, specifically `pipeline { }` as top-level expression [[4]](https://jenkins.io/doc/book/pipeline/syntax/). Do not omit parenthesis for Groovy methods because Jenkins will interpret the method as a Pipeline Step. Conversely, do omit parenthesis for Jenkins Pipeline's Steps.

#### Omit the .class suffix

In Groovy, the .class suffix is not needed. Omit the .class suffix for simplicity and better readability.

e.g. `new ExpectedException().expect(AbortException.class)`

-->  `new ExpectedException().expect(AbortException)`

#### Omit getters and setters

When declaring a field without modifier inside a Groovy bean, the Groovy compiler generates a private field and a getter and setter.

#### Do not initialize beans with named parameters

Do not initialize beans with named parameters, because it is not supported by Jenkins:

e.g. `Version javaVersion = new Version( major: 1, minor: 8)`

Initialize beans using Java syntax:

e.g. `Version javaVersion = new Version(1, 8)`

Use named parameters for Jenkins Pipeline Steps:

e.g. `sh returnStdout: true, script: command`

#### Do not use *with()* operator

The *with* operator is not yet supported by Jenkins, and it must not be used or encapsulated in a @NonCPS method.

#### Use *==* operator

Use Groovy’s `==` instead of Java `equals()` to avoid NullPointerExceptions. To compare the references of objects, instead of `==`, you should use `a.is(b)` [[1]](http://groovy-lang.org/style-guide.html).

#### Use GStrings

In Groovy, single quotes create Java Strings, and double quotes can create Java Strings or GStrings, depending if there is or not interpolation of variables [[1]](http://groovy-lang.org/style-guide.html). Using GStrings variable and string concatenation is more simple.

#### Do not use curly braces {} for variables or variable.property

For variables, or variable.property, drop the curly braces:

e.g. `echo "[INFO] ${name} version ${version.version} is installed."`

-->  `echo "[INFO] $name version $version.version is installed."`

#### Use 'single quotes' for Strings and constants

#### Use "double quotes" for GStrings

#### Use '''triple single quotes''' for multiline Strings

#### Use """triple double quotes""" for multiline GStrings

#### Use /slash/ for regular expressions

This notation avoids to double escape backslashes, making easier working with regex.

#### Use native syntax for data structures

Use the native syntax for data structures provided by Groovy like lists, maps, regex, or ranges of values.

#### Use aditional Groovy methods

Use the additional methods provided by Groovy to manipulate String, Files, Streams, Collections, and other classes.
For a complete description of all available methods, please read the GDK API [[5]](http://groovy-lang.org/groovy-dev-kit.html).

#### Use Groovy's switch

Groovy’s switch accepts any kind of type, thereby is more powerful. In this case, the use of *def* instead of a type is necessary.

#### Use alias for import

In Groovy, it is possible to assign an alias to imported packages. Use alias for imported packages to avoid the use of fully-qualified names and increase readability.

#### Use Groovy syntax to check objects

In Groovy a null, void, equal to zero, or empty object evaluates to false, and if not, evaluates to true. Instead of writing null and size checks e.g. `if (name != null && name.length > 0) {}`, use just the object `if (name) {}`.

#### Use *?.* operator

Use the safe dereference operator  *?.*, to simplify the code for accessing objects and object members safely. Using this operator, the Groovy compiler checks null objects and null object members, and returns *null* if the object or the object member is null and never throws a NullPointerException.

#### Use *?:* operator

Use Elvis operator *?:* to simplify default value validations.

#### Use *any* keyword

If the type of the exception thrown inside a try block is not important, catch any exception using the *any* keyword.

#### Use *assert*

To check parameters, return values, and more, use the assert statement.

### References

[1] Groovy's syntax: [http://groovy-lang.org/style-guide.html](http://groovy-lang.org/style-guide.html)

[2] Jenkins: [https://jenkins.io/doc/book/getting-started/](https://jenkins.io/doc/book/getting-started/)

[3] Jenkins Pipeline: [https://jenkins.io/doc/book/pipeline/](https://jenkins.io/doc/book/pipeline/)

[4] Jenkins Pipeline's syntax: [https://jenkins.io/doc/book/pipeline/syntax/](https://jenkins.io/doc/book/pipeline/syntax/)

[5] GDK: Groovy Development Kit: [http://groovy-lang.org/groovy-dev-kit.html](http://groovy-lang.org/groovy-dev-kit.html)
