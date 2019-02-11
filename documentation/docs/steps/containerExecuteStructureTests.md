# containerExecuteStructureTests

## Description

In this step [Container Structure Tests](https://github.com/GoogleContainerTools/container-structure-test) are executed.

This testing framework allows you to execute different test types against a Docker container, for example:

* Command tests (only if a Docker Deamon is available)
* File existence tests
* File content tests
* Metadata test

## Prerequisites

Test configuration is available.

## Example

```
containerExecuteStructureTests(
  script: this,
  testConfiguration: 'config.yml',
  testImage: 'node:latest'
)
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|containerCommand|no|``||
|containerShell|no|``||
|dockerImage|yes|`ppiper/container-structure-test`||
|dockerOptions|no|`-u 0 --entrypoint=''`||
|failOnError|no|`true`|`true`, `false`|
|pullImage|no||`true`, `false`|
|stashContent|no|<ul><li>`tests`</li></ul>||
|testConfiguration|no|||
|testDriver|no|||
|testImage|no|||
|testReportFilePath|no|`cst-report.json`||
|verbose|no||`true`, `false`|

Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `containerCommand`: Only for Kubernetes environments: Command which is executed to keep container alive, defaults to '/usr/bin/tail -f /dev/null'
* containerShell: Only for Kubernetes environments: Shell to be used inside container, defaults to '/bin/sh'
* dockerImage: Docker image for code execution.
* dockerOptions: Options to be passed to Docker image when starting it (only relevant for non-Kubernetes case).
* failOnError: Defines the behavior, in case tests fail.
* pullImage: Only relevant for testDriver 'docker'.
* stashContent: If specific stashes should be considered for the tests, you can pass this via this parameter.
* testConfiguration: Container structure test configuration in yml or json format. You can pass a pattern in order to execute multiple tests.
* testDriver: Container structure test driver to be used for testing, please see [https://github.com/GoogleContainerTools/container-structure-test](https://github.com/GoogleContainerTools/container-structure-test) for details.
* testImage: Image to be tested
* testReportFilePath: Path and name of the test report which will be generated
* verbose: Print more detailed information into the log.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|containerCommand||X|X|
|containerShell||X|X|
|dockerImage||X|X|
|dockerOptions||X|X|
|failOnError||X|X|
|pullImage||X|X|
|stashContent||X|X|
|testConfiguration||X|X|
|testDriver||X|X|
|testImage||X|X|
|testReportFilePath||X|X|
|verbose|X|X|X|
