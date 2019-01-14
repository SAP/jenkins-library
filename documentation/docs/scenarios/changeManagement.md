# Hybrid Application Development with Jenkins and SAP Solution Manager

Set up an agile development process which includes Jenkins CI and automatically feeds changes into SAP Solution Manager.

## Prerequisites

* You have the Java Runtime Environment 8.
* You have Jenkins 2.60.3 or higher.
* You have set up Project “Piper”. See [README](https://github.com/SAP/jenkins-library/blob/master/README.md).
* You have installed SAP Solution Manager 7.2 SP6. See [README](https://github.com/SAP/devops-cm-client/blob/master/README.md).
* You have the MTA Archive Builder 1.0.6 or any compatible version. See [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).

## Context

In many SAP development scenarios, it is vital to synchronize both backend and frontend deliveries. These deliveries are typically an SAP UI5 application and an ABAP backend from which it is served. The SAP UI5 parts are often developed agilely and use continuous integration pipelines that automatically build, test, and deploy the application.
In this scenario, we want to show how an agile development process which includes Jenkins CI can automatically feed changes into SAP Solution Manager. In SAP Solution Manager, the parts of the application stack come together and can be subject to classic change and transport management.

The basic work flow is as follows:

1. The pipeline checks the Git commit message for a change document in status `in development`. The template for the commit message looks as follows:
```
<Commit Message Header>

<Commit Message Description>

Change Document: <Your Change Document ID>
```
2. To communicate with SAP Solution Manager, the pipeline uses credentials that must be stored on Jenkins under the label `CM`.
3. The required trasport request is created on the fly. However, the change document can contain more components (for example, UI and backend components).
4. The changes of your development team trigger the Jenkins pipeline so that it builds and validates them and attaches them to the respective transport request.
5. When the development process is completed, the change document in SAP Solution Manager can be set to status `in test` and all components can be transported to the test system.

## Code

### Jenkinsfile

```
@Library('piper-library-os') _

node() {

  stage('prepare') {
      checkout scm
      setupCommonPipelineEnvironment script:this
      checkChangeInDevelopment script: this
  }

  stage('buildMta') {
      mtaBuild script: this
  }
  stage('uploadToTransportRequest') {
      transportRequestUploadFile script:this
  }
}
```

### Configuration (`.pipeline/config.yml`)

```
#Steps Specific Configuration
general:
  changeManagement:
      endpoint: 'https://<backend-system>/sap/opu/odata/sap/AI_CRM_GW_CM_CI_SRV'
credentialsId: 'CM'
     type: 'SOLMAN'
steps:
  mtaBuild:
    buildTarget: 'NEO'
    transportRequestUploadFile:
      applicationId: 'HCP'
```

## Result

This pipeline checks the git commit for a valid change ID in status “in development” on the SAP Solution Manager endpoint. It uses credentials that must be store on Jenkins under the label “CM”. Commit messages typically look as follows:

```
My change

Change description

ChanMgmtId: …
```

If the change document is OK, a transport request is created on the fly. Then, the project is built as an MTA and attached to the transport request. After that, the change is further processed in SAP Solution Manager (for example, set to status "test" or "transport").

## Parameters

For the detailed description of the relevant parameters, see:

* [checkChangeInDevelopment](https://sap.github.io/jenkins-library/steps/checkChangeInDevelopment/)
* [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/)
* [transportRequestUploadFile](https://sap.github.io/jenkins-library/steps/transportRequestUploadFile/)

## Variations

* Use the `landscape.yaml` for global landscape configuration. See [Configuration](https://sap.github.io/jenkins-library/configuration/).
* Create a transport request on the fly. See [transportRequestCreate](https://sap.github.io/jenkins-library/steps/transportRequestCreate/) and [transportRequestRelease](https://sap.github.io/jenkins-library/steps/transportRequestRelease/).
