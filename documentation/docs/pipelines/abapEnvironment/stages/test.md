# Test

This stage contains two sub stages, `ATC` and `AUnit` which can execute ATC checks and AUnit test runs in parallel on an SAP BTP ABAP environment systen.
By default this stage will not run any of the two sub stages `ATC` and `AUnit` if they are not configured. Please keep in mind that the `ATC` and `AUnit` stages need to be configured independently.
The below sections contain more information on the usage and configuration of the `ATC` and `AUnit` stages.

## ATC

In this stage, ATC checks can be executed using [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/). The step can receive software components or packages.
In case an ATC System Configuration should be used, it can be created/updated using [abapEnvironmentPushATCSystemConfig](https://sap.github.io/jenkins-library/steps/abapEnvironmentPushATCSystemConfig/).

The results are returned in the checkstyle format and can be displayed using the [Warnings Next Generation Plugin](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#warnings-next-generation-plugin). To display the results it is necessary to [extend the ATC stage via the Checkstyle/Warnings Next Generation Plugin](https://www.project-piper.io/pipelines/abapEnvironment/extensibility/#1-extend-the-atc-stage-via-the-checkstylewarnings-next-generation-plugin).

Alternatively it is possible to [extend the ATC stage to send ATC results via E-Mail](https://www.project-piper.io/pipelines/abapEnvironment/extensibility/#2-extend-the-atc-stage-to-send-atc-results-via-e-mail).

### Steps

The following steps are executed in this stage:

- [abapEnvironmentPushATCSystemConfig](../../../steps/abapEnvironmentPushATCSystemConfig.md)
- [abapEnvironmentRunATCCheck](../../../steps/abapEnvironmentRunATCCheck.md)

### Stage Parameters

There are no specifc stage parameters.

### Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

### Configuration Example

#### config.yml

In case of NOT providing an ATC System Configuration.

```yaml
general:
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'abap_system'
stages:
  ATC:
    atcConfig: 'atcConfig.yml'
```

In case of providing an ATC System Configuration.

```yaml
general:
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'myOrg'
  cfSpace: 'mySpace'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'abap_system'
stages:
  ATC:
    atcConfig: 'atcConfig.yml'
    atcSystemConfigFilePath: 'atcSystemConfig.json'
```

#### atcConfig.yml

```yaml
objectSet:
  softwarecomponents:
    - name: "/DMO/SWC"
```

#### atcSystemConfig.json

```json
{
  "conf_name": "myATCSystemConfigurationName",
  "checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
  "block_findings": "0",
  "inform_findings": "1",
  "is_default": false,
  "is_proxy_variant": false
}
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
stages:
  AUnit:
    aUnitConfig: 'aUnitConfig.yml'
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
  softwarecomponents:
  - name: Z_TEST_SC
  - name: Z_TEST_SC2
```
