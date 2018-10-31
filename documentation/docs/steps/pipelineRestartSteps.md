# pipelineRestartSteps

## Description
Support of restarting failed stages or steps in a pipeline is limited in Jenkins.

This has been documented in the [Jenkins Jira issue JENKINS-33846](https://issues.jenkins-ci.org/browse/JENKINS-33846).

For declarative pipelines there is a solution available which partially addresses this topic:
https://jenkins.io/doc/book/pipeline/running-pipelines/#restart-from-a-stage.

Nonetheless, still features are missing, so it can't be used in all cases.
The more complex Piper pipelines which share a state via [`commonPipelineEnvironment`](commonPipelineEnvironment.md) will for example not work with the standard _restart-from-stage_.

The step `pipelineRestartSteps` aims to address this gap and allows individual parts of a pipeline (e.g. a failed deployment) to be restarted.

This is done in a way that the pipeline waits for user input to restart the pipeline in case of a failure. In case this user input is not provided the pipeline stops after a timeout which can be configured.

## Prerequisites
none


## Example

Usage of pipeline step:

```groovy
pipelineRestartSteps (script: this) {
  node {
    //your steps ...
  }
}
```

!!! caution
    Use `node` inside the step. If a `node` exists outside the step context, the `input` step which is triggered in the process will block a Jenkins executor.

    In case you cannot use `node` inside this step, please choose the parameter `timeoutInSeconds` carefully!


## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|sendMail|no|`true`||
|timeoutInSeconds|no|`900`||

### Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* If `sendMail: true` the step `mailSendNotification` will be triggered in case of an error
* `timeoutInSeconds` defines the time period where the job waits for input. Default is 15 minutes. Once this time is passed the job enters state FAILED.


## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|sendMail|X|X|X|
|timeoutInSeconds|X|X|X|

## Return value
none

## Side effects
none

## Exceptions
none

