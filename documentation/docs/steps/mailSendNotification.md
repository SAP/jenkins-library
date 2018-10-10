# mailSendNotification

## Description
Sends notifications to all potential culprits of a current or previous build failure plus to fixed list of recipients.
It will attach the current build log to the email.

Notifications are sent in following cases:

* current build failed or is unstable
* current build is successful and previous build failed or was unstable

## Prerequsites
none

## Example

Usage of pipeline step:

```groovy
mailSendNotification script: this
```


## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|buildResult|no|||
|gitCommitId|no|`script.commonPipelineEnvironment.getGitCommitId()`||
|gitSshKeyCredentialsId|no|``||
|gitUrl|no|||
|notificationAttachment|no|`true`||
|notificationRecipients|no|||
|notifyCulprits|no|`true`||
|numLogLinesInBody|no|`100`||
|projectName|no|||
|wrapInNode|no|`false`||

### Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `buildResult` may be used to overrule the build result coming from `currentBuild.result`. This is for example used in the step [`pipelineRestartSteps`](pipelineRestartSteps.md)
* `gitCommitId` defines a dedicated git commitId for culprit retrieval.
* `gitUrl` and `gitCommitId` are used to retrieve culprit information.
* `gitSshKeyCredentialsId` only required if your git repository is protected. It defines the credentialsId for the git ssh credentials.
* `notificationAttachment` defines if the console log file should be attached to the notification mail.
* `notificationRecipients` defines the fixed list of recipient that always get the notification. In case you want to send the notification to the culprits only set it to an empty string `''`.

!!! note
    Multiple recipients need to be separated with the `space` character.
    In case you do not want to have any fixed recipients of the notifications leave the property empty.

* `notifyCulprits` defines if potential culprits should receive an email.
* `numLogLinesInBody` defines the number of log lines (=last lines of the log) which are included into the body of the notification email.
* `projectName` may be used to specify a different name in the email subject.
* `wrapInNode` needs to be set to `true` if step is used outside of a node context, e.g. post actions in a declarative pipeline script.


## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|buildResult||X|X|
|gitCommitId||X|X|
|gitSshKeyCredentialsId|X|X|X|
|gitUrl||X|X|
|notificationAttachment||X|X|
|notificationRecipients||X|X|
|notifyCulprits||X|X|
|numLogLinesInBody||X|X|
|projectName||X|X|
|wrapInNode||X|X|

## Return value
none

## Side effects
none

## Exceptions
none









