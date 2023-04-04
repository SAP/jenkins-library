# ${docGenStepName}

## ${docGenDescription}

!!! caution "This step is deprecated."
    Please use the [gctsExecuteABAPQualityChecks](https://www.project-piper.io/steps/gctsExecuteABAPQualityChecks/) instead.
    Don´t worry, if you´re already using the step. You can continue to use it. It will call the functions of the [gctsExecuteABAPQualityChecks](https://www.project-piper.io/steps/gctsExecuteABAPQualityChecks/) step.

## Prerequisites

* [ATC](https://help.sap.com/docs/ABAP_PLATFORM_NEW/ba879a6e2ea04d9bb94c7ccd7cdac446/62c41ad841554516bb06fb3620540e47.html) checks are enabled in transaction ATC in the ABAP systems where you want to use the step.
* [ABAP Unit tests](https://help.sap.com/docs/ABAP_PLATFORM_NEW/ba879a6e2ea04d9bb94c7ccd7cdac446/491cfd8926bc14cde10000000a42189b.html) are available for the source code that you want to check. Note: Do not execute unit tests in client 000, and not in your production client.
* [gCTS](https://help.sap.com/docs/ABAP_PLATFORM_NEW/4a368c163b08418890a406d413933ba7/f319b168e87e42149e25e13c08d002b9.html) is available and configured in the ABAP systems where you want to use the step.
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
  workspace: "${WORKSPACE}",
  queryparameters: [saml2: 'disabled']

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
