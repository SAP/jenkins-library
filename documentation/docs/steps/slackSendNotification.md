# ${docGenStepName}

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.

## ${docGenDescription}

## Prerequisites

* Installed and configured [Slack JenkinsCI integration](https://my.slack.com/services/new/jenkins-ci)
* *secret text* Jenkins credentials with the Slack token
* Installed and configured [Jenkins Slack plugin](https://github.com/jenkinsci/slack-plugin#install-instructions-for-slack)

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Usage of pipeline step:

```groovy
pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        echo "do something"
      }
    }
  }
  post {
    always {
      slackSendNotification script: this
    }
  }
}
```
