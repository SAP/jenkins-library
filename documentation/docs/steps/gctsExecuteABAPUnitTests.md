# ${docGenStepName}

## ${docGenDescription}

!!! caution "This step has been deprecated."
Please use the [gctsExecuteABAPQualityChecks](https://www.project-piper.io/steps/gctsExecuteABAPQualityChecks/) instead.
Don´t worry, if you´re already using the step. You can continue to use it. It will call the functions of the [gctsExecuteABAPQualityChecks](https://www.project-piper.io/steps/gctsExecuteABAPQualityChecks/) step.


## Prerequisites

* [ATC](https://help.sap.com/viewer/c238d694b825421f940829321ffa326a/202110.000/en-US/4ec5711c6e391014adc9fffe4e204223.html) checks are enabled in transaction ATC in the ABAP systems where you want to use the step.
* [ABAP Unit tests](https://help.sap.com/viewer/ba879a6e2ea04d9bb94c7ccd7cdac446/latest/en-US/491cfd8926bc14cde10000000a42189b.html) are available for the source code that you want to check. Note: Do not execute unit tests in client 000, and not in your production client.
* [gCTS](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/26c9c6c5a89244cb9506c253d36c3fda.html) is available and configured in the ABAP systems where you want to use the step.
* If you want to use environmental variables as parameters, for example, `GIT_COMMIT`: The [Git Plugin](https://plugins.jenkins.io/git/) is installed in Jenkins.
* The [Warnings-Next-Generation](https://plugins.jenkins.io/warnings-ng/) Plugin is installed in Jenkins.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a Jenkinsfile.

```groovy
gctsExecuteABAPUnitTests(
  script: this,
  host: 'https://abap.server.com:port',
  client: '000',
  abapCredentialsId: 'ABAPUserPasswordCredentialsId',
  repository: 'myrepo',
  scope: 'remoteChangedObjects',
  commit: "${env.GIT_COMMIT}",
  workspace: "${WORKSPACE}"

  )
```

Example configuration for the use in a yaml config file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  gctsExecuteABAPUnitTests:
    host: 'https://abap.server.com:port'
    client: '000'
    abapCredentialsId: 'ABAPUserPasswordCredentialsId'
    repository: 'myrepo'
    scope: 'remoteChangedObjects'
    commit: '38abb4814ae46b98e8e6c3e718cf1782afa9ca90'
    workspace: '/var/jenkins_home/workspace/myFirstPipeline'
```

Example configuration when you define scope: *repository* or *packages*. For these two cases you do not need to specify a *commit*.

```yaml
steps:
  <...>
  gctsExecuteABAPUnitTests:
    host: 'https://abap.server.com:port'
    client: '000'
    abapCredentialsId: 'ABAPUserPasswordCredentialsId'
    repository: 'myrepo'
    scope: 'repository'
    workspace: '/var/jenkins_home/workspace/myFirstPipeline'
```

Example configuration when you want to execute only ABAP Unit Test.

```yaml
steps:
  <...>
  gctsExecuteABAPUnitTests:
    host: 'https://abap.server.com:port'
    client: '000'
    abapCredentialsId: 'ABAPUserPasswordCredentialsId'
    repository: 'myrepo'
    atcCheck: false
    scope: 'packages'
    workspace: '/var/jenkins_home/workspace/myFirstPipeline'
```

Example configuration for the use of *recordIssue* step to make the findings visible in Jenkins interface.

```groovy
stage('ABAP Unit Tests') {
  steps{

   script{

     try{
           gctsExecuteABAPUnitTests(
              script: this,
              commit: "${env.GIT_COMMIT}",
              workspace: "${WORKSPACE}")
        }
          catch (Exception ex) {
            currentBuild.result = 'FAILURE'
            unstable(message: "${STAGE_NAME} is unstable")
             }

        }
      }
    }
stage('Results in Checkstyle') {
  steps{

     recordIssues(
          enabledForFailure: true, aggregatingResults: true,
          tools: [checkStyle(pattern: 'ATCResults.xml', reportEncoding: 'UTF8'),checkStyle(pattern: 'AUnitResults.xml', reportEncoding: 'UTF8')]
       )

      }
    }

}
```

**Note:** If you have disabled *atcCheck* or *aUnitTest*, than you also need to remove the corresponding *ATCResults.xml* or *AUnitResults.xml* from *recordIssues* step. In the example below the *atcCheck* was disabled, so *ATCResults.xml* was removed.

```groovy
recordIssues(
  enabledForFailure: true, aggregatingResults: true,
  tools: [checkStyle(pattern: 'AUnitResults.xml', reportEncoding: 'UTF8')]

)
```
