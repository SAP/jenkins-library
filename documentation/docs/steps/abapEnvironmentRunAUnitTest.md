# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

!!! Currently the Object Set configuration is limited to the usage of Multi Property Sets. Please note that other sets besides the Multi Property Set will not be included in the ABAP Unit test runs. You can see an example of the Multi Property Sets with all configurable properties. However, we strongly reccommend to only specify packages and software components like in the first two examples of the section `AUnit config file example`.

## Prerequisites

* A SAP BTP, ABAP environment system is available. On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario “SAP BTP, ABAP Environment - Software Component Test Integration (SAP_COM_0735)“. This can be done manually through the respective applications on the SAP BTP, ABAP environment system or through creating a Service Key for the system on Cloud Foundry with the parameters {“scenario_id”: “SAP_COM_0735", “type”: “basic”}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You can either provide the ABAP endpoint configuration to directly trigger an AUnit run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP BTP, ABAP environment system in Cloud Foundry that contains all the details of the ABAP endpoint to trigger an AUnit run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key, you have to provide the object set containing the objects you want to be checked in an AUnit run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile defining the pipeline.
* Make sure that the objects contained in the object set are present in the configured system in order to run the check. Please make sure that you have created or pulled the respective software components and/or packages including the test classes and objects in the SAP BTP, ABAP environment system, that should be checked.

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
The username and password to authenticate to ABAP system will then be read from the Cloud Foundry Service Key that is bound to the ABAP instance.

This can be done accordingly:

```groovy
abapEnvironmentRunAUnitTest(
    cfApiEndpoint: 'https://test.server.com',
    cfOrg: 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
    cfServiceKeyName: 'myServiceKey',
    abapCredentialsId: 'cfCredentialsId',
    aUnitConfig: 'aUnitConfig.yml',
    script: this,
)
```

To trigger the AUnit test run an AUnit config file `aUnitConfig.yml` will be needed. Check the section 'AUnit config file example' for more information.

### AUnit test run via direct ABAP endpoint configuration in Jenkinsfile

This  example triggers an AUnit run directly on the ABAP endpoint.

In order to trigger the AUnit run you have to pass the username and password for authentication to the ABAP endpoint via parameters as well as the ABAP endpoint/host. You can store the credentials in Jenkins and use the abapCredentialsId parameter to authenticate to the ABAP endpoint/host.

This must be configured as following:

```groovy
abapEnvironmentRunAUnitTest(
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    aUnitConfig: 'aUnitConfig.yml',
    script: this,
)
```

To trigger the AUnit run an AUnit config file `aUnitConfig.yml` will be needed. Check section 'AUnit config file example' for more information.

### AUnit config file example

Providing a specifc AUnit configuration is optional. If you are using a `repositories.yml` file for the `Clone` stage of the ABAP environment pipeline, a default AUnit configuration will be derived if no explicit AUnit configuration is available.

The following section contains an example of an `aUnitConfig.yml` file.
This file must be stored in the same Git folder where the `Jenkinsfile` is stored to run the pipeline. This repository containing the `Jenkinsfile` must be taken as a SCM in the Jenkins pipeline to run the pipeline.

You can specify a Multi Property Set containing multiple properties that should be checked. Each property that is specified in the Multi Property Set acts like an additional rule.
This means if you specify e.g. a Multi Property Set containing the owner and package properties that an ABAP Unit test run will be started testing all objects belonging to this owner inside of the given package. If you additionally define the Version to be `ACTIVE` for the ABAP Unit test run inside of the Multi Property Set, only objects belonging to this owner which are active inside of the package would be tested.
This must be in the same format as below examples for a `aUnitConfig.yml` file.
Note that if you want to check complete software components we recommend to use the `softwareComponent` property over the `package` property.

See below example for an `aUnitConfig.yml` file containing a minimal configuration for the software component `/DMO/SWC` to be checked:

```yaml
title: My AUnit run
context: My unit tests
objectset:
  softwarecomponents:
    - name: /DMO/SWC
```

See below example for an `aUnitConfig.yml` file with the configured options containing the package `Z_TEST_PACKAGE` to be checked:

```yaml
title: My AUnit run
context: My unit tests
objectset:
  packages:
    - name: Z_TEST_PACKAGE
```

The following example of an `aUnitConfig.yml` file containing the software component `Z_TESTSC` and shows the available options:

```yaml
title: My AUnit run
context: My unit tests
options:
  measurements: none
  scope:
    ownTests: true
    foreignTests: true
  riskLevel:
    harmless: true
    dangerous: true
    critical: true
  duration:
    short: true
    medium: true
    long: true
objectset:
  softwarecomponents:
    - name: Z_TESTSC
```

The following example of an `aUnitConfig.yml` file contains all possible properties of the Multi Property Set that can be used. Please take note that this is not the reccommended approach. If you want to check packages or software components please use the two above examples. The usage of the Multi Property Set is only reccommended for ABAP Unit tests that require these rules for the test execution. There is no official documentation on the usage of the Multi Property Set.

```yaml
title: My AUnit run
context: My unit tests
options:
  measurements: none
  scope:
    ownTests: true
    foreignTests: true
  riskLevel:
    harmless: true
    dangerous: true
    critical: true
  duration:
    short: true
    medium: true
    long: true
objectset:
  type: multiPropertySet
  multipropertyset:
    owners:
      - name: demoOwner
    softwarecomponents:
      - name: demoSoftwareComponent
    versions:
      - value: ACTIVE
    packages:
      - name: demoPackage
    objectnamepatterns:
      - value: 'ZCL_*'
    languages:
      - value: EN
    sourcesystems:
      - name: H01
    objecttypes:
      - name: CLAS
    objecttypegroups:
      - name: CLAS
    releasestates:
      - value: RELEASED
    applicationcomponents:
      - name: demoApplicationComponent
    transportlayers:
      - name: H01
```
