# Test

This stage contains two sub stages, `ATC` and `AUnit` which can execute ATC checks and AUnit test runs in parallel on an SAP BTP ABAP environment systen.
By default this stage will not run any of the two sub stages `ATC` and `AUnit` if they are not configured. Please keep in mind that the `ATC` and `AUnit` stages need to be configured independently.
The below sections contain more information on the usage and configuration of the `ATC` and `AUnit` stages.

## ATC

In this stage, ATC checks can be executed using abapEnvironmentRunATCCheck. The step can receive software components or packages. The results are returned in the checkstyle format. With the use of a pipeline extension, quality gates can be configured (see [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/) or the "Extensions" section in the [configuration](../configuration.md)).

### Steps

The following steps are executed in this stage:

- [abapEnvironmentRunATCCheck](../../../steps/abapEnvironmentRunATCCheck.md)

### Stage Parameters

There are no specifc stage parameters.

### Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage..

### Configuration Example

#### config.yml

```yaml
general:
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'abap_system'
  cfServiceKeyName: 'JENKINS_SAP_COM_0510'
stages:
  ATC:
    atcConfig: 'atcConfig.yml'
```

#### atcConfig.yml

```yaml
atcobjects:
  softwarecomponent:
    - name: "/DMO/SWC"
```

#### ATC.groovy

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  recordIssues tools: [checkStyle(pattern: '**/ATCResults.xml')], qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

Please note, this file belongs in the extensions folder: `.pipeline/extensions/ATC.groovy`

## AUnit

This stage will trigger an AUnit test run an on SAP BTP ABAP Environment system using the abapEnvironmentRunAUnitTest step.

### Steps

The following steps are executed in this stage:

- [abapEnvironmentRunAUnitTest](../../../steps/abapEnvironmentRunAUnitTest.md)

### Stage Parameters

There are no specifc stage parameters.

### Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

### Configuration Example

#### config.yml

```yaml
general:
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'abap_system'
  cfServiceKeyName: 'JENKINS_SAP_COM_0510'
stages:
  AUnit:
    aunitConfig: 'aunitConfig.yml'
```

#### aunitConfig.yml

If you want to test complete software components please specify the `aUnitConfig.yml` file like in below example configuration. This configuration will test the software components `Z_TEST_SC` and `Z_TEST_SC2`:

```yaml
title: My AUnit run
context: AUnit test run
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
objectSet:
  softwarecomponent:
  - name: Z_TEST_SC
  - name: Z_TEST_SC2
```
