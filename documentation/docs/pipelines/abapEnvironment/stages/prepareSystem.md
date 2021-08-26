# Prepare System

In this stage, the SAP BTP, ABAP environment system is created. This is done with the `abapEnvironmentCreateSystem` step. After the system creation, the Communication Arrangement [SAP_COM_0510](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/b04a9ae412894725a2fc539bfb1ca055.html) (SAP BTP, ABAP Environment - Software Component Test Integration) is created using the step `cloudFoundryCreateServiceKey`. With the creation of the Communication Arrangement, a User and Password is created on the SAP BTP, ABAP environment system for the APIs that are used in the following stages.

## Steps

The following steps are executed in this stage:

- [abapEnvironmentCreateSystem](../../../steps/abapEnvironmentCreateSystem.md)
- [cloudFoundryCreateServiceKey](../../../steps/cloudFoundryCreateServiceKey.md)

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
  Prepare System:
    cfService: 'abap'
    cfServicePlan: 'standard'
    abapSystemAdminEmail: 'user@example.com'
    abapSystemDescription: 'ABAP Environment Q System'
    abapSystemIsDevelopmentAllowed: false
    abapSystemID: 'H02'
    abapSystemSizeOfPersistence: 4
    abapSystemSizeOfRuntime: 1
    cfServiceKeyConfig: 'serviceKey.json'
```

### serviceKey.json

```json
{
  "scenario_id": "SAP_COM_0510",
  "type": "basic"
}
```
