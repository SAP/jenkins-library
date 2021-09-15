# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* A SAP BTP, ABAP environment system is available. On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario “SAP BTP, ABAP Environment - Software Component Test Integration (SAP_COM_0735)“. This can be done manually through the respective applications on the SAP BTP, ABAP environment system or through creating a service key for the system on Cloud Foundry with the parameters {“scenario_id”: “SAP_COM_0735", “type”: “basic”}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You can either provide the ABAP endpoint configuration to directly trigger an AUnit run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP BTP, ABAP environment system in Cloud Foundry that contains all the details of the ABAP endpoint to trigger an ATC run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key, you have to provide the Object Set containing the Objects you want to be checked in an AUnit run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile defining the pipeline.
* Make sure that the Objects contained in the Object Set are present in the configured system in order to run the check. Please make sure that you have created or pulled the respective software components and/or Packages including the Test Classes and Objects in the SAP BTP, ABAP environment system, that should be checked.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### AUnit test run via direct ABAP endpoint configuration in Jenkinsfile

This  example triggers an AUnit test run directly on the ABAP endpoint.

In order to trigger the AUnit test run you have to pass the username and password for authentication to the ABAP endpoint via parameters as well as the ABAP endpoint/host. You can store the credentials in Jenkins and use the abapCredentialsId parameter to authenticate to the ABAP endpoint/host.

This must be configured as following:

```groovy
abapEnvironmentRunAUnitTest(
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    aUnitConfig: 'aUnitConfig.yml',
    script: this,
)
```

To trigger the AUnit test run an AUnit config file `aUnitConfig.yml` will be needed. Check the section 'AUnit config file example' for more information.

### AUnit test run via Cloud Foundry Service Key example in Jenkinsfile

The following example triggers an AUnit test run via reading the Service Key of an ABAP instance in Cloud Foundry.

You can store the credentials in Jenkins and use the cfCredentialsId parameter to authenticate to Cloud Foundry.
The username and password to authenticate to ABAP system will then be read from the Cloud Foundry service key that is bound to the ABAP instance.

This can be done accordingly:

```groovy
abapEnvironmentRunAUnitTest(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
    cfServiceKeyName: 'myServiceKey',
    abapCredentialsId: 'cfCredentialsId',
    aUnitConfig: 'aUnitConfig.yml',
    script: this,
)
```

To trigger the AUnit test run an AUnit config file `aUnitConfig.yml` will be needed. Check the section 'AUnit config file example' for more information.

### ATC run via direct ABAP endpoint configuration in Jenkinsfile

This  example triggers an ATC run directly on the ABAP endpoint.

In order to trigger the ATC run you have to pass the username and password for authentication to the ABAP endpoint via parameters as well as the ABAP endpoint/host. You can store the credentials in Jenkins and use the abapCredentialsId parameter to authenticate to the ABAP endpoint/host.

This must be configured as following:

```groovy
abapEnvironmentRunATCCheck(
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    aUnitConfig: 'aUnitConfig.yml',
    script: this,
)
```

To trigger the ATC run an ATC config file `aUnitConfig.yml` will be needed. Check section 'ATC config file example' for more information.

### AUnit config file example

The following section contains an example of an `aUnitConfig.yml` file.
This file must be stored in the same Git folder where the `Jenkinsfile` is stored to run the pipeline. This folder must be taken as a SCM in the Jenkins pipeline to run the pipeline.

You can specify an Object Set containing the Objects that should be checked. These can be for example packages, classes or transport requests containing test classes that can be executed. This must be in the same format as below example for a `aUnitConfig.yml` file.
Note that if you specify a package inside a packageSet to be checked for each package that has to be checked you can configure if you want the subpackages to be included in checks or not.

See below example for an `aUnitConfig.yml` file containing a package to be checked:

```yaml
title: My AUnit run
context: My unit tests
options:
  measurements: none
  scope:
    owntests: true
    foreigntests: true
  riskLevel:
    harmless: true
    dangerous: true
    critical: true
  duration:
    short: true
    medium: true
    long: true
objectset:
  - type: unionSet
    set:
      - type: packageSet
        package:
          - name: my_package
            includesubpackages: false
```

The following example of an `aUnitConfig.yml` file containing one class and one interface to be checked:

```yaml
title: My AUnit run
context: My unit tests
options:
  measurements: none
  scope:
    owntests: true
    foreigntests: true
  riskLevel:
    harmless: true
    dangerous: true
    critical: true
  duration:
    short: true
    medium: true
    long: true
objectset:
  - type: unionSet
    set:
      - type: flatObjectSet
        object:          
        - name: my_class
          type: CLAS
        - name: my_interface
          type: INTF
```