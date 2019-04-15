# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* Installed and configured [Jenkins Slack plugin](https://github.com/jenkinsci/slack-plugin).

## ${docGenParameters}

## ${docGenConfiguration}

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
