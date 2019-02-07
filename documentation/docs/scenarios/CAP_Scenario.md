# Build and Deploy Applications with Jenkins and the SAP Cloud Application Programming Model

Set up a basic continuous delivery process for developing applications according to the SAP Cloud Application Programming Model.

## Prerequisites

* You have an account in the Cloud Foundry environment. See [Accounts](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/8ed4a705efa0431b910056c0acdbf377.html).
* You have downloaded and installed the Cloud Foundry command line interface (CLI). See [Download and Install the Cloud Foundry Command Line Interface](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/afc3f643ec6942a283daad6cdf1b4936.html).
* You have installed the multi-target application plug-in for the Cloud Foundry command line interface. See [Install the Multi-Target Application Plug-in in the Cloud Foundry Environment](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/27f3af39c2584d4ea8c15ba8c282fd75.html).
* You have installed the Java Runtime Environment 8.
* You have installed Jenkins 2.60.3 or higher.
* You have set up Project “Piper”. See [README](https://github.com/SAP/jenkins-library/blob/master/README.md).
* You have installed the Multi-Target Application (MTA) Archive Builder 1.0.6 or newer. See [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).
* You have installed Node.js including node and npm. See [Node.js](https://nodejs.org/en/download/).

## Context

The Application Programming Model for SAP Cloud Platform is an end-to-end best practice guide for developing applications on SAP Cloud Platform and provides a supportive set of APIs, languages, and libraries. For more information about the SAP Cloud Application Programming Model, see [Working with the SAP Cloud Application Programming Model](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/00823f91779d4d42aa29a498e0535cdf.html).

In this scenario, we want to show how to implement a basic continuous delivery process for developing applications according to this programming model with the help of project "Piper" on Jenkins.

**Note**: This document does not provide you with a fully developed continuous integration (CI) and continuous delivery (CD) pipeline but describes a basic scenario to implement automatisation, which you can adapt and enrich according to your specific needs.

## Example

### Jenkinsfile

```groovy
@Library('piper-library-os') _

node(){  
   stage('Prepare')   {  
      deleteDir()
      checkout scm
      setupCommonPipelineEnvironment script:this
   }  
   
   stage('Build')   {  
      mtaBuild script:this, buildTarget:'CF'
   }   
   
   stage('Deploy')   {  
      cloudFoundryDeploy script:this, deployTool:'mtaDeployPlugin'
   }
}
```

### Configuration (`.pipeline/config.yml`)

```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'CF'
      apiEndpoint: '<CF Endpoint>'
      org: '<CF Organization>'
      space: '<CF Space>'
```

### Parameters

For the detailed description of the relevant parameters, see:

* [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/)
* [cloudFoundryDeploy](https://sap.github.io/jenkins-library/steps/cloudFoundryDeploy/)
