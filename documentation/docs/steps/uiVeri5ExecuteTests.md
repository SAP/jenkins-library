# uiVeri5ExecuteTests

## Description

With this step [UIVeri5](https://github.com/SAP/ui5-uiveri5) tests can be executed.

UIVeri5 describes following benefits on its GitHub page:

* Automatic synchronization with UI5 app rendering so there is no need to add waits and sleeps to your test. Tests are reliable by design.
* Tests are written in synchronous manner, no callbacks, no promise chaining so are really simple to write and maintain.
* Full power of webdriverjs, protractor and jasmine - deferred selectors, custom matchers, custom locators.
* Control locators (OPA5 declarative matchers) allow locating and interacting with UI5 controls.
* Does not depend on testability support in applications - works with autorefreshing views, resizing elements, animated transitions.
* Declarative authentications - authentication flow over OAuth2 providers, etc.
* Console operation, CI ready, fully configurable, no need for java (comming soon) or IDE.
* Covers full ui5 browser matrix - Chrome,Firefox,IE,Edge,Safari,iOS,Android.
* Open-source, modify to suite your specific neeeds.

!!! note "Browser Matrix"
    With this step and the underlying Docker image ([selenium/standalone-chrome](https://github.com/SeleniumHQ/docker-selenium/tree/master/StandaloneChrome)) only Chrome tests are possible.

    Testing of further browsers can be done with using a custom Docker image.

## Prerequisites

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `dockerEnvVars` | no |  |  |
| `dockerImage` | no |  |  |
| `dockerWorkspace` | no |  |  |
| `failOnError` | no |  | `true`, `false` |
| `gitBranch` | no |  |  |
| `gitSshKeyCredentialsId` | no |  | Jenkins credentialId |
| `installCommand` | no | `npm install @ui5/uiveri5 --global --quiet` |  |
| `runCommand` | no | `uiveri5 --seleniumAddress='http://${config.seleniumHost}:${config.seleniumPort}/wd/hub'` |  |
| `script` | yes |  |  |
| `seleniumHost` | no |  |  |
| `seleniumPort` | no | `4444` |  |
| `sidecarEnvVars` | no |  |  |
| `sidecarImage` | no |  |  |
| `stashContent` | no | `[buildDescriptor, tests]` |  |
| `testOptions` | no |  |  |
| `testRepository` | no |  |  |

* `dockerEnvVars` - A map of environment variables to set in the container, e.g. [http_proxy:'proxy:8080'].
* `dockerImage` - The name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
* `dockerWorkspace` - Only relevant for Kubernetes case: Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
* `failOnError` - With `failOnError` the behavior in case tests fail can be defined.
* `gitBranch` - In case a `testRepository` is provided the branch in this repository can be specified with `gitBranch`.
* `gitSshKeyCredentialsId` - In case a `testRepository` is provided and it is protected, access credentials (as Jenkins credentials) can be provided with `gitSshKeyCredentialsId`. **Note: In case of using a protected repository, `testRepository` should include the ssh link to the repository.**
* `installCommand` - The command that is executed to install the test tool.
* `runCommand` - The command that is executed to start the tests.
* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in `script: this`. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.
* `seleniumHost` - The host of the selenium hub, this is set automatically to `localhost` in a Kubernetes environment (determined by the `ON_K8S` environment variable) of to `selenium` in any other case. The value is only needed for the `runCommand`.
* `seleniumPort` - The port of the selenium hub. The value is only needed for the `runCommand`.
* `sidecarEnvVars` - A map of environment variables to set in the sidecar container, similar to `dockerEnvVars`.
* `sidecarImage` - The name of the docker image of the sidecar container. If empty, no sidecar container is started.
* `stashContent` - If specific stashes should be considered for the tests, their names need to be passed via the parameter `stashContent`.
* `testOptions` - This allows to set specific options for the UIVeri5 execution. Details can be found [in the UIVeri5 documentation](https://github.com/SAP/ui5-uiveri5/blob/master/docs/config/config.md#configuration).
* `testRepository` - With `testRepository` the tests can be loaded from another reposirory.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections of the config.yml the configuration is possible:

| parameter | general | step | stage |
|-----------|---------|------|-------|
| `dockerEnvVars` |  | X | X |
| `dockerImage` |  | X | X |
| `dockerWorkspace` |  | X | X |
| `failOnError` |  | X | X |
| `gitBranch` |  | X | X |
| `gitSshKeyCredentialsId` | X | X | X |
| `installCommand` |  | X | X |
| `runCommand` |  | X | X |
| `script` |  |  |  |
| `seleniumHost` |  | X | X |
| `seleniumPort` |  | X | X |
| `sidecarEnvVars` |  | X | X |
| `sidecarImage` |  | X | X |
| `stashContent` |  | X | X |
| `testOptions` |  | X | X |
| `testRepository` |  | X | X |

## Exceptions

## Examples
