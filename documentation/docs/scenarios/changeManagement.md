# Hybrid Application Development with Jenkins and SAP Solution Manager

Set up an agile development process which includes Jenkins CI and automatically feeds changes into SAP Solution Manager.

## Prerequisites

* You meet the requirements for Project “Piper”. See [Requirements](https://github.com/SAP/jenkins-library/blob/master/README.md#requirements).
* You have set up Project “Piper”. See [Download and Installation](https://github.com/SAP/jenkins-library/blob/master/README.md#download-and-installation).
* You have installed SAP Solution Manager 7.2 SP6. See [README](https://github.com/SAP/devops-cm-client/blob/master/README.md).
* You meet the prerequisites for the mtaBuild. See [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/).

## Context

In many SAP development scenarios, it is vital to synchronize both backend and frontend deliveries. These deliveries are typically an SAP UI5 application and an ABAP backend from which it is served. The SAP UI5 parts are often developed agilely and use continuous integration pipelines that automatically build, test, and deploy the application.
In this scenario, we want to show how an agile development process which includes Jenkins CI can automatically feed changes into SAP Solution Manager. In SAP Solution Manager, the parts of the application stack come together and can be subject to classic change and transport management.

The basic work flow is as follows:

1. Check SAP Solution Manager for a change document.
2. Make sure that in SAP Solution Manager, there are transport requests for the components that are part of the delivery (for example, parts of SAP Cloud Platform and S/4HANA).
3. Your development team makes changes which trigger the Jenkins pipeline.
4. Jenkins builds and validates the changes and attaches them to the respective transport request.
5. When the development team has finished, the change document in SAP Solution Manager is set to status “in test” and all components are transported to the Q system.

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
      git:
          from: 'HEAD~1'
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
