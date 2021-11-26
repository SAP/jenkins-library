# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* ATC checks are enabled in transaction ATC in the ABAP systems where you want to use the step (https://help.sap.com/viewer/c238d694b825421f940829321ffa326a/202110.000/en-US/4ec5711c6e391014adc9fffe4e204223.html).
* gCTS is available and configured in the ABAP systems where you want to use the step (https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/26c9c6c5a89244cb9506c253d36c3fda.html).
* The Static Analysis Warning plug-in (Warnings Next Generation Plugin) is installed in Jenkins(https://www.jenkins.io/doc/pipeline/steps/warnings-ng/).



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
  scope: 'LOCAL_CHANGED_OBJECTS',
  commitId: "${GIT_COMMIT}",
  jenkinsWorkspace: "${WORKSPACE}"

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
    scope: 'LOCAL_CHANGED_OBJECTS'
    commitId: '0123456789abcdefghijkl'
    jenkinsWorkspace: '/var/jenkins_home/workspace/myfirstpipeline'
```
