# Pipeline for SAP UI5/Fiori On SAP Cloud Platform

This is a so called scenario step. Scenario steps are aggregations of several steps implementing a simple, but complete pipeline. This should make simple scenarios easy to set up. Your `Jenkinsfile` can be as simple as:

```groovy
@Library('piper-lib-os') _

<scenario>Pipeline script: this

```

## Description
This steps builds an SAP UI5 or Fiori based application using MTA and deploys the build result into an SAP Cloud Platform (Neo) account. This scenario wraps the [mtaBuild](mtaBuild.md) and [neoDeploy](neoDeploy.mta) steps.


![This pipeline in Jenkins Blue Ocean](images/pipeline.jpg)

## Prerequisites

#### General Prerequisites
- project "Piper" requires a Jenkins (2.x) with pipeline plugins to run
- project "Piper" needs to be [registered in Jenkins](https://github.com/SAP/jenkins-library/blob/master/README.md) as a Global Shared Library

More specifically, the steps included in this scenario might require additional files in your project and execution environment on your Jenkins. 

#### Prerequisites for the MTA Build

* A docker image meeting the following requirements:
  * **SAP MTA Archive Builder** - can be downloaded from [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).
  * **Java 8 or compatible version** - necessary to run the `mta.jar` file.
  * **NodeJS** - the MTA Builder requires `node` and `npm` to build the project.

For more information please check the documentation for the [MTA build](mtaBuild.md).

#### Prerequisites for the Deployment to SAP Cloud Platform

* **SAP CP account** - the account to where the application is deployed.
* **SAP CP user for deployment** - a user with deployment permissions in the given account.
* **Jenkins credentials for deployment** - must be configured in Jenkins credentials with a dedicated Id.

![Jenkins credentials configuration](../images/neo_credentials.png)

* **Neo Java Web** - can be downloaded from [Maven Central](http://central.maven.org/maven2/com/sap/cloud/neo-java-web-sdk/).
* **Java 8 or compatible version** - needed by the *Neo-Java-Web-SDK*

For more information please check the documentation for the [deployment](neoDeploy.md).

# Parameters

#### Parameters for the MTA Build

| parameter        | mandatory | default                                                | possible values    |
| -----------------|-----------|--------------------------------------------------------|--------------------|
| `script`         | yes       |                                                        |                    |
| `dockerImage`    | yes       |                                                        |                    |
| `buildTarget`    | yes       | `'NEO'`                                                | 'CF', 'NEO', 'XSA' |
| `mtaJarLocation` | no        | `'mta.jar'`                                        |                    |

For the full list of parameters please check the documentation for the [MTA build](mtaBuild.md).

#### Parameters for the Deployment to SAP Cloud Platform

| parameter          | mandatory | default                       | possible values                                 |
| -------------------|-----------|-------------------------------|-------------------------------------------------|
| `deployMode`       | yes       | `'mta'`                       | `'mta'`, `'warParams'`, `'warPropertiesFile'`   |
| `script`           | yes       |                               |                                                 |

For the full list of parameters please check the documentation for the [deployment](neoDeploy.md).

# Step Configuration

Please refer to our configurations documentation and the documentation for the individual steps:

* [General configuration](configuration)
* [MTA build configuration](mtaBuild.md)
* [Deployment configuration](neoDeploy.md)


## Example

#### Jenkinsfile
```groovy
@Library('piper-lib-os') _

fioriOnCloudPlatformPipeline script:this
```

#### .pipeline/config.yml

```yaml
steps:
  mtaBuild:
    buildTarget: 'NEO'
  neoDeploy:
    neoCredentialsId: 'NEO_DEPLOY'
    account: 'your-account-id'
```

# Project Template Files

The following template files needs to be provided and adjusted on project level:

#### `.npmrc`

The [`.npmrc`](documentation/docs/scenarios/ui5-sap-cp/files/.npmrc)
  contains a reference to the SAP NPM registry: `@sap:registry https://npm.sap.com` that is required to fetch dependencies to build the application.

#### `mta.yaml`

The [`mta.yaml`](documentation/docs/scenarios/ui5-sap-cp/files/mta.yaml) controls the behavior of the mta toolset. Place the file in your application root folder and adjust the values in brackets with your data.

#### `package.json`

The [package.json](documentation/docs/scenarios/ui5-sap-cp/files/package.json) fetches the (dev-)dependencies that are required to build. Add the lines to your existing `package.json` file.


#### `Gruntfile.js`
[Gruntfile.js](documentation/docs/scenarios/ui5-sap-cp/files/Gruntfile.js) controls the grunt build. By default these tasks are executed: `clean`, `build`, `lint`.
