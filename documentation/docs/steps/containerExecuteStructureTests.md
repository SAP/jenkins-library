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
|failOnError|no|`true`||
|pullImage|no|||
|stashContent|no|<ul><li>`tests`</li></ul>||
|testConfiguration|no|||
|testDriver|no|||
|testImage|no|||
|testReportFilePath|no|`cst-report.json`||
|verbose|no|||

Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.



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


