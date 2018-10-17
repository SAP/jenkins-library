# gaugeExecuteTests

## Description
In this step Gauge ([getgauge.io](http:getgauge.io)) acceptance tests are executed.
Using Gauge it will be possible to have a three-tier test layout:
* Acceptance Criteria
* Test implemenation layer
* Application driver layer

This layout is propagated by Jez Humble and Dave Farley in their book "Continuous Delivery" as a way to create maintainable acceptance test suites (see "Continuous Delivery", p. 190ff).

Using Gauge it is possible to write test specifications in [Markdown syntax](http://daringfireball.net/projects/markdown/syntax) and therefore allow e.g. product owners to write the relevant acceptance test specifications. At the same time it allows the developer to implement the steps described in the specification in her development environment.

You can use the sample projects of Gauge, for example: https://github.com/getgauge/gauge-mvn-archetypes

!!! note "Make sure to run against a Selenium Hub configuration"
    In the test example of _gauge-archetype-selenium_ please make sure to allow it to run against a Selenium hub:

    Please extend DriverFactory.java for example in following way:

    ``` java
    String hubUrl = System.getenv("HUB_URL");
    //when running on a Docker deamon (and not using Kubernetes plugin), Docker images will be linked
    //in this case hubUrl will be http://selenium:4444/wd/hub due to the linking of the containers
    hubUrl = (hubUrl == null) ? "http://localhost:4444/wd/hub" : hubUrl;
    Capabilities chromeCapabilities = DesiredCapabilities.chrome();
    System.out.println("Running on Selenium Hub: " + hubUrl);
    return new RemoteWebDriver(new URL(hubUrl), chromeCapabilities);
    ```

## Prerequsites

none

## Example

Pipeline step:
```groovy
gaugeExecuteTests script: this, testServerUrl: 'http://test.url'
```


## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|buildTool|no|`maven`||
|dockerEnvVars|no|`[HUB:TRUE, HUB_URL:http://localhost:4444/wd/hub]`||
|dockerImage|no|buildTool=`maven`: `maven:3.5-jdk-8`<br />buildTool=`npm`: `node:8-stretch`<br />||
|dockerName|no|buildTool=`maven`: `maven`<br />buildTool=`npm`: `npm`<br />||
|dockerWorkspace|no|buildTool=`maven`: ``<br />buildTool=`npm`: `/home/node`<br />||
|failOnError|no|`false`||
|gitBranch|no|||
|gitSshKeyCredentialsId|no|``||
|installCommand|no|`curl -SsL https://downloads.gauge.org/stable | sh -s -- --location=$HOME/bin/gauge`||
|languageRunner|no|buildTool=`maven`: `java`<br />buildTool=`npm`: `js`<br />||
|runCommand|no|buildTool=`maven`: `mvn test-compile gauge:execute`<br />buildTool=`npm`: `gauge run`<br />||
|stashContent|no|<ul><li>`buildDescriptor`</li><li>`tests`</li></ul>||
|testOptions|no|buildTool=`maven`: `-DspecsDir=specs`<br />buildTool=`npm`: `specs`<br />||
|testRepository|no|||
|testServerUrl|no|||


Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `buildTool` defines the build tool to be used for the test execution.
* `dockerEnvVars`, see step [dockerExecute](dockerExecute.md)
* `dockerImage`, see step [dockerExecute](dockerExecute.md)
* `dockerName`, see step [dockerExecute](dockerExecute.md)
* `dockerWorkspace`, see step [dockerExecute](dockerExecute.md)
* With `failOnError` you can define the behavior, in case tests fail. When this is set to `true` test results cannot be recorded using the `publishTestResults` step afterwards.
* `installCommand` defines the command for installing Gauge. In case the `dockerImage` already contains Gauge it can be set to empty: ``.
* `languageRunner` defines the Gauge language runner to be used.
* `runCommand` defines the command which is used for executing Gauge.
* If specific stashes should be considered for the tests, you can pass this via parameter `stashContent`
* `testOptions` allows to set specific options for the Gauge execution. Details can be found for example [in the Gauge Maven plugin documentation](https://github.com/getgauge/gauge-maven-plugin#executing-specs)
* In case the test implementation is stored in a different repository than the code itself, you can define the repository containing the tests using parameter `testRepository` and if required `gitBranch` (for a different branch than master) and `gitSshKeyCredentialsId` (for protected repositories). For protected repositories the `testRepository` needs to contain the ssh git url.
* `testServerUrl` is passed as environment variable `TARGET_SERVER_URL` to the test execution. Tests running against the system should read the host information from this environment variable in order to be infrastructure agnostic.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|buildTool||X|X|
|dockerEnvVars||X|X|
|dockerImage||X|X|
|dockerName||X|X|
|dockerWorkspace||X|X|
|failOnError||X|X|
|gitBranch||X|X|
|gitSshKeyCredentialsId||X|X|
|stashContent||X|X|
|testOptions||X|X|
|testRepository||X|X|
|testServerUrl||X|X|
