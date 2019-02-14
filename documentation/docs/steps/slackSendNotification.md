# slackSendNotification

## Description

Sends notifications to the Slack channel about the build status.

Notification contains:

* Build status;
* Repo Owner;
* Repo Name;
* Branch Name;
* Jenkins Build Number;
* Jenkins Build URL.

## Prerequisites

Installed and configured [Jenkins Slack plugin](https://github.com/jenkinsci/slack-plugin).

## Example

Usage of pipeline step:

```groovy
try {
  stage('..') {..}
  stage('..') {..}
  stage('..') {..}
  currentBuild.result = 'SUCCESS'
} catch (Throwable err) {
  currentBuild.result = 'FAILURE'
  throw err
} finally {
  stage('report') {
    slackSendNotification script: this
  }
}
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|baseUrl|no|||
|channel|no|||
|color|no|`${buildStatus == 'SUCCESS'?'#008000':'#E60000'}`||
|credentialsId|no|||
|message|no|||

### Details

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `baseUrl` allows overriding the Slack Plugin Integration Base Url specified in the global configuration.
* `color` defines the message color.
* If `channel` is defined another than the default channel will be used.
* `credentialsId` defines the Jenkins credentialId which holds the Slack token
* With parameter `message` a custom message can be defined which is sent into the Slack channel.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|baseUrl||X|X|
|channel||X|X|
|color||X|X|
|credentialsId||X|X|
|message||X|X|
