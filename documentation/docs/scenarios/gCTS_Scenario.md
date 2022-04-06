# Set up a Pipeline-Based ABAP Development and Testing Process Using Git-Enabled Change and Transport System.

For current information about gCTS, see SAP Note [Central Note for Git-enabled Change and Transport System (gCTS)](https://launchpad.support.sap.com/#/notes/2821718)

## Introduction

[Git-enabled Change & Transport System (gCTS)](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/f319b168e87e42149e25e13c08d002b9.html) enables you to manage your ABAP change and transport management processes using Git as an external version management system. It allows you to set up continuous integration processes for ABAP development.  

This scenario explains how to use a pipeline to deploy a commit to a test system, and execute [ABAP unit tests](https://help.sap.com/viewer/ba879a6e2ea04d9bb94c7ccd7cdac446/latest/en-US/491cfd8926bc14cde10000000a42189b.html) and [ATC (ABAP Test Cockpit)](https://help.sap.com/viewer/ba879a6e2ea04d9bb94c7ccd7cdac446/latest/en-US/62c41ad841554516bb06fb3620540e47.html) checks in the test system. For each new commit that arrives in the remote repository, the pipeline executes the following Piper steps in the test system:     

1. [gctsDeploy](../../steps/gctsDeploy/) step: Deploys the commit on the test system.
2. [gctsExecuteABAPQualityChecks](../../steps/gctsExecuteABAPQualityChecks/) step: Executes ABAP unit tests and ATC checks for the ABAP development objects of the commit.
   - If the result of the testing is *success*, the pipeline finishes.
   - If the result of the testing is *error*, a rollback is executed (see next step).
3. [gctsRollback](../../steps/gctsRollback/) step: Executes a rollback to the previous commit. You can check the cause of the errors using the [Warnings Next Generation Plugin](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#warnings-next-generation-plugin) in Jenkins.


## Prerequisites

- You have configured Git-Enabled Change and Transport System, and you use it for your ABAP development. See
 [Configuring Git-enabled Change & Transport System (gCTS)](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/26c9c6c5a89244cb9506c253d36c3fda.html) 

- You have a Git repository on a Git server, such as GitHub, or GitLab.  
    The Git repository is usually created as part of the gCTS configuration. It is used to store your ABAP developments.
    You can use this Git repository also for the pipeline configuration. (Jenkinsfile)
    The repository used for the pipeline configuration needs to be accessed by the Jenkins instance. If the repository is password-protected, the user and password (or access token) should be stored in the Jenkins Credentials Store (**Manage Jenkins** > **Manage Credentials**).

- You have at least two ABAP systems with a version SAP S/4HANA 2020 or higher. You need one development system that you use to push objects to the Git repository, and a test system on which you run the pipeline. You have created and cloned the Git repository on all systems, on the development system with the *Development* role, and in the others with the *Provided* role.

- You have enabled [ATC](https://help.sap.com/viewer/ba879a6e2ea04d9bb94c7ccd7cdac446/latest/en-US/62c41ad841554516bb06fb3620540e47.html) checks in transaction ATC in the test system.

- You have access to a Jenkins instance including the [Warnings-Next-Generation Plugin](https://plugins.jenkins.io/warnings-ng/). The plug-in must be installed separately. It is required to view the results of the testing after the pipeline has run.  
    For the gCTS scenario, we recommend that you use the [Custom Jenkins setup](https://www.project-piper.io/infrastructure/customjenkins/) even though it is possible to run the gCTS scenario with [Piper´s CX server](https://www.project-piper.io/infrastructure/overview/).

- You have set up a suitable Jenkins instance as described under [Getting Started with Project "Piper"](https://www.project-piper.io/guidedtour/) under *Create Your First Pipeline*.

- The user that is used for the execution of the pipeline must have the credentials entered in gCTS as described in the gCTS documentation under [Set User-Specific Authentication](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/3431ebd6fbf241778cd60587e7b5dc3e.html).


## Process

The process is as follows:  

You create or change ABAP objects in the development system. When you release the transport request, the objects are pushed to the remote repository in a new commit. The pipeline is triggered by the new commit. The pipeline can be started manually in Jenkins, or automatically when the new commit arrives in the Git repository (by setting a webhook in GitHub). For more information about webhooks, see [Creating webhooks](https://docs.github.com/en/developers/webhooks-and-events/webhooks/creating-webhooks).

The pipeline deploys the new commit on the test system and executes ATC checks and ABAP Unit tests for the objects of the commit depending on the specified object scope. In the sample scenario, this is `localChangedObjects`, which means, for all objects that were changed by the last activity in the local repository.

If the checks don´t issue any errors, the pipeline finishes. The test system remains on the new commit, and you can continue your testing activities, for example, using manual tests.

If the checks result in any errors, the pipeline executes a rollback to the last active commit in the test system. You can display errors and warnings of the checks in the [Warnings-Next-Generation Plugin](https://plugins.jenkins.io/warnings-ng/). For more information about the mapping of the check errors to errors displayed in Jenkins, see the description of the following parameters in the [gctsExecuteABAPQualityChecks](https://www.project-piper.io/steps/gctsExecuteABAPQualityChecks/) step:
- Severities of ABAP Unit test results: [aUnitTest](../../steps/gctsExecuteABAPQualityChecks/#aunittest) parameter
- Priorities of ATC check results: [atcCheck](../../steps/gctsExecuteABAPQualityChecks/#atccheck) parameter

After analyzing the errors, you can correct the issues in the development system. Once you release the new transport request, the pipeline is triggered again.

The following image shows the steps involved when the checks finish successfully:

![Process: Deploy Git repository on local system and execute tests - Tests are successful](../images/gctscheckSuccessful.png "Process: Deploy and execute tests: Success")  
**Image: Build and Deploy Process in Jenkins**


The following image shows the steps involved when the checks result in warnings or errors:

![Process: Deploy Git repository on local system and execute tests - Tests are not successful](../images/gctscheckNotSuccessful.png "Process: Deploy and execute tests: Success")  
**Image: Build and Deploy Process in Jenkins**


## Example

### Jenkinsfile

Following the convention for pipeline definitions, use a Jenkinsfile, which resides in the root directory of your Git repository. In the following sample Jenkinsfile, all configuration information required for setting up the gCTS scenario is defined in the Jenkinsfile. As an alternative, you can use a `.pipeline/config.yml` file to define parts of your configuration. See the [Configuration](#Configuration) example below.   

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
          branch 'main'
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
          vSID: '<vSID>')

      }
    }

    stage('gctsExecuteABAPQualityChecks') {
      when {
        anyOf {
          branch 'main'
        }
      }
      steps {
        script {
          try {
          gctsExecuteABAPQualityChecks(
          script: this,
          host: HOST,
          client: CLIENT,
          abapCredentialsId: DEMOCREDS,
          repository: REPO,
          scope: 'localChangedObjects',
          commit: "${env.GIT_COMMIT}",
          workspace: "${WORKSPACE}")
        } catch (Exception ex) {
          currentBuild.result = 'FAILURE'
          unstable(message: "${STAGE_NAME} is unstable")
        }

      }
    }
  }

stage('Results in Checkstyle') {
  when {
      anyOf {
        branch 'main'
      }
    }
  steps{

     recordIssues(
          enabledForFailure: true, aggregatingResults: true,
          tools: [checkStyle(pattern: 'ATCResults.xml', reportEncoding: 'UTF8'),checkStyle(pattern: 'AUnitResults.xml', reportEncoding: 'UTF8')]
       )

      }
    }   
stage('Rollback') {
            when {
              expression {
                currentBuild.result == 'FAILURE'
              }
            }
            steps {
              gctsRollback(
                script: this,
                host: HOST,
                client: CLIENT,
                abapCredentialsId: DEMOCREDS,
                repository: REPO
          )

      }
    }
  }

}
```   

### Configuration  (`.pipeline/config.yml`)

If you don´t define all configuration parameters directly in the Jenkinsfile (as done in the sample Jenkinsfile above), you can use an additional `config.yml` file for configuration. For general information about configuration in "Piper" projects, see [Configuration](https://www.project-piper.io/configuration/).

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
    gctsExecuteABAPQualityChecks(
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

For a detailed description of the relevant parameters, see:
- [gctsDeploy](../../steps/gctsDeploy/)
- [gctsExecuteABAPQualityChecks](../../steps/gctsExecuteABAPQualityChecks/)
- [gctsRollback](../../steps/gctsRollback/)

## Troubleshooting

If you encounter an issue with the pipeline itself, please open an issue in [GitHub](https://github.com/SAP/jenkins-library/issues) and add the label `gcts` to the issue.
