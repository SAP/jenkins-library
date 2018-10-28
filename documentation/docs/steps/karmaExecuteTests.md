# karmaExecuteTests

## Description

In this step the ([Karma test runner](http://karma-runner.github.io)) is executed.

The step is using the `seleniumExecuteTest` step which spins up two containers in a Docker network:
 - a Selenium container (`selenium/standalone-chrome`)
 - a NodeJS container (`node:8-stretch`)
In the Docker network, the containers can be referenced by the values provided in `dockerName` and `sidecarName`, the default values are `karma` and `selenium`. These values must be used in the test configuration ([Karma `hostname`](https://karma-runner.github.io/1.0/config/configuration-file.html) and [WebDriver `hostname`](https://github.com/karma-runner/karma-webdriver-launcher#usage)).

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
|dockerEnvVars|no|||
|dockerImage|no|`node:8-stretch`||
|dockerName|no|`karma`||
|dockerWorkspace|no|`/home/node`||
|failOnError|no|||
|installCommand|no|`npm install --quiet`||
|modules|no|`['.']`||
|runCommand|no|`npm run karma`||
|sidecarEnvVars|no|||
|sidecarImage|no|||
|sidecarName|no|||
|sidecarVolumeBind|no|||
|stashContent|no|||

* `<parameter>` - Detailed description of each parameter.

## Step configuration

* `<parameter>`

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
