# karmaExecuteTests

## Description

In this step the ([Karma test runner](http://karma-runner.github.io)) is executed.

The step is using the `seleniumExecuteTest` step to spins up two containers in a Docker network:

- a Selenium/Chrome container (`selenium/standalone-chrome`)
- a NodeJS container (`node:8-stretch`)

In the Docker network, the containers can be referenced by the values provided in `dockerName` and `sidecarName`, the default values are `karma` and `selenium`. These values must be used in the `hostname` properties of the test configuration ([Karma](https://karma-runner.github.io/1.0/config/configuration-file.html) and [WebDriver](https://github.com/karma-runner/karma-webdriver-launcher#usage)).

!!! note
    In a Kubernetes environment, the containers both need to be referenced with `localhost`.

## Prerequisites

- **running Karma tests** - have a NPM module with running tests executed with Karma
- **configured WebDriver** - have the [`karma-webdriver-launcher`](https://github.com/karma-runner/karma-webdriver-launcher) package installed and a custom, WebDriver-based browser configured in Karma

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|containerPortMappings|no|`[node:8-stretch: [[containerPort: 9876, hostPort: 9876]]]`||
|dockerEnvVars|no|`[ NO_PROXY: 'localhost,karma,$NO_PROXY', no_proxy: 'localhost,karma,$no_proxy']`||
|dockerImage|no|`node:8-stretch`||
|dockerName|no|`karma`||
|dockerWorkspace|no|`/home/node`||
|failOnError|no|||
|installCommand|no|`npm install --quiet`||
|modules|no|`['.']`||
|runCommand|no|`npm run karma`||
|sidecarEnvVars|no|`[ NO_PROXY: 'localhost,selenium,$NO_PROXY', no_proxy: 'localhost,selenium,$no_proxy']`||
|sidecarImage|no|||
|sidecarName|no|||
|sidecarVolumeBind|no|||
|stashContent|no|`['buildDescriptor', 'tests']`||

- `script` - defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
- `containerPortMappings` - see step [dockerExecute](dockerExecute.md)
- `dockerEnvVars` - see step [dockerExecute](dockerExecute.md)
- `dockerImage` - see step [dockerExecute](dockerExecute.md)
- `dockerName` - see step [dockerExecute](dockerExecute.md)
- `dockerWorkspace` - see step [dockerExecute](dockerExecute.md)
- `failOnError` - see step [seleniumExecuteTests](seleniumExecuteTests.md)
- `installCommand` - the command that is executed to install dependencies
- `modules` - define the paths of the modules to execute tests on
- `runCommand` - the command that is executed to start the tests
- `sidecarEnvVars` - see step [dockerExecute](dockerExecute.md)
- `sidecarImage` - see step [dockerExecute](dockerExecute.md)
- `sidecarName` - see step [dockerExecute](dockerExecute.md)
- `sidecarVolumeBind` - see step [dockerExecute](dockerExecute.md)
- `stashContent` - pass specific stashed that should be considered for the tests

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|---------|------|-------|
|script||||
|containerPortMappings|X|X|X|
|dockerEnvVars|X|X|X|
|dockerImage|X|X|X|
|dockerName|X|X|X|
|dockerWorkspace|X|X|X|
|failOnError|X|X|X|
|installCommand|X|X|X|
|modules|X|X|X|
|runCommand|X|X|X|
|sidecarEnvVars|X|X|X|
|sidecarImage|X|X|X|
|sidecarName|X|X|X|
|sidecarVolumeBind|X|X|X|
|stashContent|X|X|X|

## Return value

none

## Side effects

Step uses `seleniumExecuteTest` & `dockerExecute` inside.

## Exceptions

none

## Example

```groovy
karmaExecuteTests script: this, modules: ['./shoppinglist', './catalog']
```
