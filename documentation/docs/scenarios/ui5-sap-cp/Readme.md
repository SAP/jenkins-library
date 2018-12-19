# Pipeline for SAP UI5/Fiori On SAP Cloud Platform

This is a so called scenario step. Scenario steps are aggregations of several steps implementing a complete pipeline. This makes typical scenarios easy to set up.

## Description

This steps builds an SAP UI5 or Fiori based application using MTA and deploys the build result into an SAP Cloud Platform (Neo) account. This scenario wraps the [mtaBuild](../../steps/mtaBuild.md) and [neoDeploy](../../steps/neoDeploy.md) steps.

![This pipeline in Jenkins Blue Ocean](images/pipeline.jpg)

To implement this you can use the following `Jenkinsfile` and `config.yml`. See below for prerequisites and configuration options.

### Jenkinsfile

The convention for pipeline definitions is to use a `Jenkinsfile` that resides in the root directory of your development sources.

```groovy
@Library('piper-lib-os') _

fioriOnCloudPlatformPipeline script:this
```

### .pipeline/config.yml

This is a basic configuration example that is also located in the project's sources.

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

## Prerequisites

### General Prerequisites

- Project "Piper" requires a Jenkins (2.x) with pipeline plugins to run.
- Project "Piper" needs to be [registered in Jenkins](https://github.com/SAP/jenkins-library/blob/master/README.md) as a Global Shared Library.
- **Java 8 or compatible version** is required for build and deployment tools.

### Prerequisites for the MTA Build

A docker image meeting the following requirements:
- **SAP MTA Archive Builder** - can be downloaded from [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).
- **NodeJS** - the MTA Builder requires `node` and `npm` to build the project.

For more information please check the documentation for the [MTA build](../../steps/mtaBuild.md).

### Prerequisites for the Deployment to SAP Cloud Platform

- **SAP CP account** - the account to where the application is deployed.
- **SAP CP user for deployment** - a user with deployment permissions in the given account.
- **Jenkins credentials for deployment** - must be configured in Jenkins credentials with a dedicated ID.
- **Neo Java Web** - can be downloaded from [Maven Central](http://central.maven.org/maven2/com/sap/cloud/neo-java-web-sdk/).

For more information please check the documentation for the [deployment](../../steps/neoDeploy.md).

### Prerequisites in Your Project

The steps included in this scenario require additional files in your project and execution environment on your Jenkins. 

The following template files needs to be provided and adjusted on project level:

| File Name | Comment |
|-----|-----|
| [`.npmrc`](https://github.com/marcusholl/jenkins-library/tree/pr/scenarioUI5SAPCP/documentation/docs/scenarios/ui5-sap-cp/files/.npmrc) | Contains a reference to the SAP NPM registry: `@sap:registry https://npm.sap.com` that is required to fetch dependencies to build the application. Place it in your project's root directoy. |
| [`mta.yaml`](https://github.com/marcusholl/jenkins-library/tree/pr/scenarioUI5SAPCP/documentation/docs/scenarios/ui5-sap-cp/files/mta.yaml) | Controls the behavior of the mta toolset. Place the file in your application root folder and adjust the values in brackets with your data. |
| [`package.json`](https://github.com/marcusholl/jenkins-library/tree/pr/scenarioUI5SAPCP/documentation/docs/scenarios/ui5-sap-cp/files/package.json) | Lists the (dev-)dependencies that are required to build. Add the content to your existing `package.json` file. |
| [`Gruntfile.js`](https://github.com/marcusholl/jenkins-library/tree/pr/scenarioUI5SAPCP/documentation/docs/scenarios/ui5-sap-cp/files/Gruntfile.js) | Controls the grunt build. By default these tasks are executed: `clean`, `build`, `lint`. Place it in your project's root directoy. |

## Step Configuration

The configuration must be stored in the `.pipeline/config.yml`.

### Configuration for the MTA Build

| Parameter        | Description    |
| -----------------|----------------|
| `buildTarget`    | The target platform to which the mtar can be deployed, possible values: `CF`, `NEO`, `XSA` |
| `mtaJarLocation` | The location of the SAP Multitarget Application Archive Builder jar file, including file name and extension. |

For the full list of configuration options please check the documentation for the [MTA build](../../steps/mtaBuild.md).

### Configuration for the Deployment to SAP Cloud Platform

| Parameter          | Description |
| -------------------|-------------|
| `account`           | The SAP Cloud Platform _account_ to deploy to. |
| `host`           |  The SAP Cloud Platform _host_ to deploy to.. |
| `neoCredentialsId` | The Jenkins credentials (not the password!) containing user and password used for SAP CP deployment. |
| `neoHome`           | The path to the `neo-java-web-sdk` tool used for the deployment. |

For the full list of configuration options please check the documentation for the [deployment](../../steps/neoDeploy.md).

For detailed information, please refer to our configurations documentation and the documentation for the individual steps:

- [General configuration](../../configuration)
- [MTA build configuration](../../steps/mtaBuild.md)
- [Deployment configuration](../../steps/neoDeploy.md)
