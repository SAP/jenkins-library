# Build and Deploy SAP Cloud Application Programming Model Applications

In this scenario, we will setup a CI/CD Pipeline for a SAP Cloud Application Programming Model (CAP) project.

## Prerequisites

* You have the SAP Cloud Application Programming Model command line tool (cds-dk) installed: See [Get Started](https://cap.cloud.sap/docs/get-started/#local-setup).
* You have setup a suitable Jenkins instance as described in [Guided Tour](../guidedtour.md)

## Context

The Application Programming Model for SAP Business Technology Platform (SAP BTP) is an end-to-end best practice guide for developing applications on SAP BTP and provides a supportive set of APIs, languages, and libraries.
For more information about the SAP Cloud Application Programming Model, visit its [documentation](https://cap.cloud.sap/docs/about/).

## Getting started

To get started, generate a project using the SAP Cloud Application Programming Model command line tools:

```
cds init bookshop --add java,mta,samples,hana
```

Alternatively you can also reuse an existing project. To include support for continuous delivery, you can execute the command `cds add pipeline` in the directory in which you have created your project:

```
cd bookshop
cds add pipeline
```

This will generate a project which already includes a `Jenkinsfile`, and a `.pipeline/config.yml` file.

Now, you'll need to push the code to a git repository.
This is required because the pipeline gets your code via git.
This might be GitHub, or any other cloud or on-premise git solution you have in your company.

Afterwards you can connect your Jenkins instance to your git repository and let it build the project.

## Legacy documentation

If your project is not based on the _SAP Business Application Studio_ WebIDE template, you could either migrate your code to comply with the structure which is described [here](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/doc/pipeline/build-tools.md#sap-cloud-application-programming-model--mta), or you can use a self built pipeline, as described in this section.

### Prerequisites

* You have an account on SAP Business Technology Platform in the Cloud Foundry environment. See [Accounts](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/8ed4a705efa0431b910056c0acdbf377.html).
* You have downloaded and installed the Cloud Foundry command line interface (CLI). See [Download and Install the Cloud Foundry Command Line Interface](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/afc3f643ec6942a283daad6cdf1b4936.html).
* You have installed the multitarget application (MTA) plug-in for the Cloud Foundry command line interface. See [Install the Multitarget Application Plug-in in the Cloud Foundry Environment](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/27f3af39c2584d4ea8c15ba8c282fd75.html).
* You have installed the Java Runtime Environment 8.
* You have installed Jenkins 2.60.3 or higher.
* You have set up Project “Piper”. See [README](https://github.com/SAP/jenkins-library/blob/master/README.md).
* You have installed the multitarget application archive builder 1.0.6 or newer. See [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).
* You have installed Node.js including node and npm. See [Node.js](https://nodejs.org/en/download/).

### Context

The Application Programming Model for SAP Business Technology Platform is an end-to-end best practice guide for developing applications on SAP BTP and provides a supportive set of APIs, languages, and libraries. For more information about the SAP Cloud Application Programming Model, see [Working with the SAP Cloud Application Programming Model](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/00823f91779d4d42aa29a498e0535cdf.html).

In this scenario, we want to show how to implement a basic continuous delivery process for developing applications according to this programming model with the help of project "Piper" on Jenkins. This basic scenario can be adapted and enriched according to your specific needs.

### Example

#### Jenkinsfile

```groovy
@Library('piper-lib-os') _

node(){
  stage('Prepare')   {
      deleteDir()
      checkout scm
      setupCommonPipelineEnvironment script:this
  }

  stage('Build')   {
      mtaBuild script:this
  }

  stage('Deploy')   {
      cloudFoundryDeploy script:this, deployTool:'mtaDeployPlugin'
  }
}
```

#### Configuration (`.pipeline/config.yml`)

```yaml
steps:
  mtaBuild:
    buildTarget: 'CF'
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'CF'
      apiEndpoint: '<CF Endpoint>'
      org: '<CF Organization>'
      space: '<CF Space>'
```

#### Parameters

For the detailed description of the relevant parameters, see:

* [mtaBuild](../steps/mtaBuild.md)
* [cloudFoundryDeploy](../steps/cloudFoundryDeploy.md)
