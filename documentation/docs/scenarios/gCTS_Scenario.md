# Set up a Pipeline-Based ABAP Development and Testing Process Using Git-Enabled Change and Transport System.

For current information about gCTS, see SAP Note [Central Note for Git-enabled Change and Transport System (gCTS)](https://launchpad.support.sap.com/#/notes/2821718)

## Introduction

[Git-enabled Change & Transport System (gCTS)](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/f319b168e87e42149e25e13c08d002b9.html) enables you to manage your ABAP change and transport management processes using Git as an external version management system. It allows you to set up continuous integration processes for ABAP development.
This scenario explains how to use a pipeline to deploy a commit to a test system, and execute ABAP unit tests and ATC (ABAP Test Cockpit) checks in the test system. In detail, this scenario covers the following steps:    
For each new commit that arrives in the remote repository, the pipeline executes the following Piper steps in the test system:
1. [gctsDeploy](https://www.project-piper.io/steps/gctsDeploy/): Deploys the commit on the test system.
2. [gctsExecuteABAPUnitTests](https://www.project-piper.io/steps/gctsExecuteABAPUnitTests/): Executes ABAP unit tests and ATC checks for the ABAP development objects of the commit.
- If the result of the testing is success, the pipeline finishes.  
- If the result of the testing is error, a rollback to the previous commit is executed. You can check the results of the testing using the [Warnings Next Generation Plugin](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#warnings-next-generation-plugin) in Jenkins.

## Prerequisites

- You have configured Git-Enabled Change and Transport System, and you use it for your ABAP development. See
 [Configuring Git-enabled Change & Transport System (gCTS)](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/26c9c6c5a89244cb9506c253d36c3fda.html) 
- You have a Git repository on a Git server, such as GitHub, or GitLab.  
The Git repository is usually created as part of the gCTS configuration. It is used to store your ABAP developments.
You can use this Git repository also for the pipeline configuration.  
The repository used for the pipeline configuration needs to be accessed by the Jenkins instance. If the repository is password protected, the user and password (or access token) should be stored in the Jenkins Credentials Store (Manage Jenkins  &rightarrow; Manage Credentials).
- You have at least two ABAP systems with a version SAP S/4HANA 2020 or higher. You need one development system that you use to push objects to the Git repository, and a test system on which you run the pipeline. You have created and cloned the Git repository on all systems, on the development system with the *Development* role, and in the others with the *Provided* role.
- You have enabled [ATC](https://help.sap.com/viewer/c238d694b825421f940829321ffa326a/latest/en-US/4ec5711c6e391014adc9fffe4e204223.html) checks in transaction ATC in the test system.
- You have access to a Jenkins instance including the [Warnings-Next-Generation Plugin](https://plugins.jenkins.io/warnings-ng/). The plug-in must be installed separately. It is required to view the results of the testing after the pipeline has run.  
For the gCTS scenario, we recommend that you use the [Custom Jenkins setup](https://www.project-piper.io/infrastructure/customjenkins/) even though it is possible to run the gCTS scenario with [Piper´s CX server](https://www.project-piper.io/infrastructure/overview/).
- You have set up a suitable Jenkins instance as described under [Getting Started with Project "Piper"](https://www.project-piper.io/guidedtour/) under *Create Your First Pipeline*.
- The user that is used for the execution of the pipeline must have the credentials entered in gCTS as described in the gCTS documentation under [Set User-Specific Authentication](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/3431ebd6fbf241778cd60587e7b5dc3e.html).


## Process

The process is as follows:  
You create or change ABAP objects in the development system. When you release the transport request, the objects are pushed to the remote repository in a new commit. The pipeline is triggered by the new commit. It can be started manually in Jenkins, or automatically when the new commit arrives in the Git repository (by setting a webhook in GitHub).  
The following image shows the library steps involved when the tests are run successfully:
![Process: Deploy Git repository on local system and execute tests - Tests are successful](../images/checkSuccessful.png "Process: Deploy and execute tests: Success")  

The following image shows the library steps involved when the tests result in an error:
![Process: Deploy Git repository on local system and execute tests - Tests are not successful](../images/checkNotSuccessful.png "Process: Deploy and execute tests: Success")



## Example

### Jenkinsfile

If you use the pipeline of the following code snippet, you only have to configure it in the .pipeline/config.yml.

Following the convention for pipeline definitions, use a Jenkinsfile, which resides in the root directory of your development sources.

```groovy
@Library(['piper-lib-os']) _
pipeline {
  agent any
  options {
    disableConcurrentBuilds()
  }

  environment {
    DEMOCREDS = 'ABAPUserPasswordCredentialsId'
    HOST = 'https://<host of the ABAP system>:<port>'
    CLIENT = '000'
    REPO = '<repository name>'
    REPO_URL = "<URL of the remote Git Repository>"
  }

  stages {
    stage('gCTS Deploy') {
      when {
        anyOf {
          branch 'master'
        }
      }
      steps {
        gctsDeploy(
          script: this,
          host: HOST,
          client: CLIENT,
          abapCredentialsId: DEMOCREDS,
          repository: REPO,
          remoteRepositoryURL: REPO_URL,
          role: 'SOURCE',
          vSID: 'ABC')

      }
    }

    stage('gctsExecuteABAPUnitTests') {
      when {
        anyOf {
          branch 'main'
        }
      }
      steps {
        gctsExecuteABAPUnitTests(
          script: this,
          host: HOST,
          client: CLIENT,
          abapCredentialsId: DEMOCREDS,
          repository: REPO,
          scope: 'localChangedObjects',
          commit: "${GIT_COMMIT}",
          workspace: "${WORKSPACE}")

      }
    }
  }
}
stage('ABAP Unit Tests') {
  steps{

   script{

     try{
           gctsExecuteABAPUnitTests(
              script: this,
              commit: "${GIT_COMMIT}",
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
