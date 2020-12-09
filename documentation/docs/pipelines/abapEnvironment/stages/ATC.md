# ATC

In this stage, ATC checks can be executed using abapEnvironmentRunATCCheck. The step can receive software components or packages. The results are returned in the checkstyle format. With the use of a pipeline extension, quality gates can be configured (see [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/) or the "Extensions" section in the [configuration](../configuration.md)).

## Steps

The following steps are executed in this stage:

- [abapEnvironmentRunATCCheck](../../../steps/abapEnvironmentRunATCCheck.md)

## Stage Parameters

There are no specifc stage parameters.

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

## Configuration Example

### config.yml

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

### atcConfig.yml

```yaml
atcobjects:
  softwarecomponent:
    - name: "/DMO/SWC"
```

### ATC.groovy

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
