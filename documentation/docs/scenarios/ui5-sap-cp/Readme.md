# Build and Deploy SAP UI5 or SAP Fiori Applications on SAP Cloud Platform with Jenkins

Build an application based on SAP UI5 or SAP Fiori with Jenkins and deploy the build result into an SAP Cloud Platform account in the Neo environment.

## Prerequisites

* You have installed the Java Runtime Environment 8.
* You have installed Jenkins 2.60.3 or higher.
* You have set up Project “Piper”. See [README](https://github.com/SAP/jenkins-library/blob/master/README.md).
* You have installed the Multi-Target Application (MTA) Archive Builder 1.0.6 or newer. See [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).
* You have installed Node.js including node and npm. See [Node.js](https://nodejs.org/en/download/).
* You have installed the SAP Cloud Platform Neo Environment SDK. See [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).


### Project Prerequisites

This scenario requires additional files in your project and in the execution environment on your Jenkins instance. 


On the project level, provide and adjust the following template:

| File Name | Description | Position |
|-----|-----|-----|
| [`.npmrc`](https://github.com/SAP/jenkins-library/blob/master/documentation/docs/scenarios/ui5-sap-cp/files/.npmrc) | This file contains a reference to the SAP NPM registry (`@sap:registry https://npm.sap.com`), which is required to fetch the dependencies required to build the application. | Place the `.npmrc` file in the root directory of your project. |
| [`mta.yaml`](https://github.com/SAP/jenkins-library/blob/master/documentation/docs/scenarios/ui5-sap-cp/files/mta.yaml) | This file controls the behavior of the MTA toolset. | Place the `mta.yaml` file in your application root folder and adjust the values in brackets with your data. |
| [`package.json`](https://github.com/SAP/jenkins-library/blob/master/documentation/docs/scenarios/ui5-sap-cp/files/package.json) | This file lists the required development dependencies for the build. | Add the content of the `package.json` file to your existing `package.json` file. |
| [`Gruntfile.js`](https://github.com/SAP/jenkins-library/blob/master/documentation/docs/scenarios/ui5-sap-cp/files/Gruntfile.js) | This file controls the grunt build. By default the tasks `clean`, `build`, and `lint` are executed. | Place the `Gruntfile.js` in the root directory of your project. |


## Context

This scenario combines various different steps to create a complete pipeline.


In this scenario, we want to show how to build an application based on SAP UI5 or SAP Fiori by using the multi-target application (MTA) concept and how to deploy the build result into an SAP Cloud Platform account in the Neo environment. This document comprises the [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/) and the [neoDeploy](https://sap.github.io/jenkins-library/steps/neoDeploy/) steps.

![This pipeline in Jenkins Blue Ocean](images/pipeline.jpg)
###### Screenshot: Build and Deploy Process in Jenkins

## Example

### Jenkinsfile

Following the convention for pipeline definitions, use a `Jenkinsfile` which resides in the root directory of your development sources.

```groovy
@Library('piper-lib-os') _

fioriOnCloudPlatformPipeline script:this
```

### Configuration (`.pipeline/config.yml`)

This is a basic configuration example, which is also located in the sources of the project.

```yaml
steps:
  mtaBuild:
    buildTarget: 'NEO'
    mtaJarLocation: '/opt/sap/mta.jar'
  neoDeploy:
    neoCredentialsId: 'NEO_DEPLOY'
    neoHome: '/opt/sap/neo-sdk/'
    account: 'your-account-id'
    host: 'hana.ondemand.com'
```

#### Configuration for the MTA Build

| Parameter        | Description    |
| -----------------|----------------|
| `buildTarget`    | The target platform to which the mtar can be deployed. Possible values are: `CF`, `NEO`, `XSA` |
| `mtaJarLocation` | The location of the multi-target application archive builder jar file, including file name and extension. |


#### Configuration for the Deployment to SAP Cloud Platform

| Parameter          | Description |
| -------------------|-------------|
| `account`           | The SAP Cloud Platform account to deploy to. |
| `host`           |  The SAP Cloud Platform host to deploy to. |
| `neoCredentialsId` | The Jenkins credentials that contain the user and password which are used for the deployment on SAP Cloud Platform. |
| `neoHome`           | The path to the `neo-java-web-sdk` tool that is used for the deployment. |


### Parameters

For the detailed description of the relevant parameters, see:

* [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/)
* [neoDeploy](https://sap.github.io/jenkins-library/steps/neoDeploy/)
