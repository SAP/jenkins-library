# seleniumExecuteTests

## Description

Enables UI test execution with Selenium in a sidecar container.

The step executes a closure (see example below) connecting to a sidecar container with a Selenium Server.

When executing in a

* local Docker environment, please make sure to set Selenium host to **`selenium`** in your tests.
* Kubernetes environment, plese make sure to set Seleniums host to **`localhost`** in your tests.

!!! note "Proxy Environments"
    If work in an environment containing a proxy, please make sure that `localhost`/`selenium` is added to your proxy exclusion list, e.g. via environment variable `NO_PROXY` & `no_proxy`. You can pass those via parameters `dockerEnvVars` and `sidecarEnvVars` directly to the containers if required.

## Prerequisites

none

## Example

```groovy
seleniumExecuteTests (script: this) {
    git url: 'https://github.wdf.sap.corp/xxxxx/WebDriverIOTest.git'
    sh '''npm install
        node index.js'''
}
```

### Example test using WebdriverIO

Example based on http://webdriver.io/guide/getstarted/modes.html and http://webdriver.io/guide.html

#### Configuration for Local Docker Environment

```js
var webdriverio = require('webdriverio');
var options = {
    host: 'selenium',
    port: 4444,
    desiredCapabilities: {
        browserName: 'chrome'
    }
};
```

#### Configuration for Kubernetes Environment

```js
var webdriverio = require('webdriverio');
var options = {
    host: 'localhost',
    port: 4444,
    desiredCapabilities: {
        browserName: 'chrome'
    }
};
```

#### Test Code (index.js)

```js
// ToDo: add configuration from above

webdriverio
    .remote(options)
    .init()
    .url('http://www.google.com')
    .getTitle().then(function(title) {
        console.log('Title was: ' + title);
    })
    .end()
    .catch(function(err) {
        console.log(err);
    });
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|buildTool|no|`npm`|`maven`, `npm`|
|containerPortMappings|no|`[selenium/standalone-chrome:[[containerPort:4444, hostPort:4444]]]`||
|dockerEnvVars|no|||
|dockerImage|no|buildTool=`maven`: `maven:3.5-jdk-8`<br />buildTool=`npm`: `node:8-stretch`<br />||
|dockerName|no|buildTool=`maven`: `maven`<br />buildTool=`npm`: `npm`<br />||
|dockerWorkspace|no|buildTool=`maven`: ``<br />buildTool=`npm`: `/home/node`<br />||
|failOnError|no|`true`||
|gitBranch|no|||
|gitSshKeyCredentialsId|no|``||
|sidecarEnvVars|no|||
|sidecarImage|no|`selenium/standalone-chrome`||
|sidecarName|no|`selenium`||
|sidecarVolumeBind|no|`[/dev/shm:/dev/shm]`||
|stashContent|no|<ul><li>`tests`</li></ul>||
|testRepository|no|||

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `buildTool` defines the build tool to be used for the test execution.
* `containerPortMappings`, see step [dockerExecute](dockerExecute.md)
* `dockerEnvVars`, see step [dockerExecute](dockerExecute.md)
* `dockerImage`, see step [dockerExecute](dockerExecute.md)
* `dockerName`, see step [dockerExecute](dockerExecute.md)
* `dockerWorkspace`, see step [dockerExecute](dockerExecute.md)
* `failOnError` specifies if the step should fail in case the execution of the body of this step fails.
* `sidecarEnvVars`, see step [dockerExecute](dockerExecute.md)
* `sidecarImage`, see step [dockerExecute](dockerExecute.md)
* `sidecarName`, see step [dockerExecute](dockerExecute.md)
* `sidecarVolumeBind`, see step [dockerExecute](dockerExecute.md)
* If specific stashes should be considered for the tests, you can pass this via parameter `stashContent`
* In case the test implementation is stored in a different repository than the code itself, you can define the repository containing the tests using parameter `testRepository` and if required `gitBranch` (for a different branch than master) and `gitSshKeyCredentialsId` (for protected repositories). For protected repositories the testRepository needs to contain the ssh git url.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|buildTool||X|X|
|containerPortMappings|X|X|X|
|dockerEnvVars|X|X|X|
|dockerImage|X|X|X|
|dockerName|X|X|X|
|dockerWorkspace|X|X|X|
|failOnError|X|X|X|
|gitBranch|X|X|X|
|gitSshKeyCredentialsId|X|X|X|
|sidecarEnvVars|X|X|X|
|sidecarImage|X|X|X|
|sidecarName|X|X|X|
|sidecarVolumeBind|X|X|X|
|stashContent|X|X|X|
|testRepository|X|X|X|

## Return value

none

## Side effects

none

## Exceptions

none
