# Set up a pipeline-based ABAP development and testing process using Git-Enabled Change and Transport System.

!!! caution "Current limitations"
    For current information about gCTS, see SAP Note [Central Note for Git-enabled Change and Transport System (gCTS)](https://launchpad.support.sap.com/#/notes/2821718)

## Introduction

[Git-enabled Change & Transport System (gCTS)](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/f319b168e87e42149e25e13c08d002b9.html) enables you to manage your ABAP change and transport management processes using Git as an external version management system. It allows you to set up continuous integration processes for ABAP development.
This scenario explains how to use a pipeline to deploy a commit to a test system, and execute ABAP unit tests and ATC (ABAP Test Cockpit) checks in the test system. In detail, this scenario covers the following steps:    
For each new commit that arrives in the remote repository, the pipeline executes the following Piper steps in the test system:
1. [gctsDeploy](https://www.project-piper.io/steps/gctsDeploy/): Deploys the commit on the test system.
2. [gctsExecuteABAPUnitTests](https://www.project-piper.io/steps/gctsExecuteABAPUnitTests/): Executes ABAP unit tests and ATC (ABAP Test Cockpit) checks for the ABAP development objects of the commit.
- If the result of the testing is success, the pipeline finishes.  
- If the result of the testing is error, a rollback to the previous commit is executed. You can check the results of the testing using the Warnings Next Generation Plugin in Jenkins.

## Prerequisites

- You have configured Git-Enabled Change and Transport System, and you use it for your ABAP development. See
 [Configuring Git-enabled Change & Transport System (gCTS)](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/26c9c6c5a89244cb9506c253d36c3fda.html) 
- You have a Git repository on a Git server, such as GitHub, or GitLab.  
The Git repository is usually created as part of the gCTS configuration. It is used to store your ABAP developments.
You can use this Git repository also for the pipeline configuration.  
The repository used for the pipeline configuration needs to be accessed by the Jenkins instance. If the repository is password protected, the user and password (or access token) should be stored in the Jenkins Credentials Store (Manage Jenkins  &rightarrow; Manage Credentials).
- You have at least two ABAP systems with a version SAP S/4HANA 2020 or higher. You need one development system that you use to push objects to the Git repository, and a test system on which you run the pipeline.
- You have enabled [ATC](https://help.sap.com/viewer/c238d694b825421f940829321ffa326a/latest/en-US/4ec5711c6e391014adc9fffe4e204223.html) checks in transaction ATC in the test system.
- You have access to a Jenkins instance including the [Warnings-Next-Generation Plugin](https://plugins.jenkins.io/warnings-ng/).
- For the gCTS scenario, we recommend that you use the [Custom Jenkins setup](https://www.project-piper.io/infrastructure/customjenkins/). Even though it is possible to run the gCTS scenario with [Piper´s CX server](https://www.project-piper.io/infrastructure/overview/).


##Process
[Process: Deploy and execute tests](../images/checkSuccessful.png "Process: Deploy and execute tests")

##example

### Jenkinsfile
If you use the pipeline of the following code snippet, you only have to configure it in the .pipeline/config.yml.

Following the convention for pipeline definitions, use a Jenkinsfile, which resides in the root directory of your development sources.

```groovy
@Library('piper-lib-os') _

piperPipeline script:this
```

### Configuration (`.pipeline/config.yml`)

This is a basic configuration example, which is also located in the sources of the project.

```yaml
steps:
  gctsDeploy(
      script: this,
      host: 'https://abap.server.com:port',
      client: '000',
      abapCredentialsId: 'ABAPUserPasswordCredentialsId',
      repository: 'myrepo',
      remoteRepositoryURL: "https://remote.repository.url.com",
      role: 'SOURCE',
      vSID: 'ABC',
      branch: 'feature1',
      commit: '95952ec',
      scope: 'LASTACTION',
      rollback: true,
      configuration: [VCS_AUTOMATIC_PULL: 'FALSE',VCS_AUTOMATIC_PUSH: 'FALSE',CLIENT_VCS_LOGLVL: 'debug']
    )
    gctsExecuteABAPUnitTests(
      script: this,
      host: 'https://abap.server.com:port',
      client: '000',
      abapCredentialsId: 'ABAPUserPasswordCredentialsId',
      repository: 'myrepo',
      scope: 'localChangedObjects',
      commit: "${GIT_COMMIT}",
      workspace: "${WORKSPACE}"

      )
```

### Parameters

For a detailed description of the relevant parameters, see [gctsDeploy](../../steps/gctsDeploy/) and [gctsExecuteABAPUnitTests](../../steps/gctsExecuteABAPUnitTests/).

## Troubleshooting

If you encounter an issue with the pipeline itself, please open an issue in [GitHub](https://github.com/SAP/jenkins-library/issues).
