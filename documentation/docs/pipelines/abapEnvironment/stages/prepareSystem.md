# Prepare System

In this stage, the SAP BTP, ABAP environment system is created. This is done with the `abapEnvironmentCreateSystem` step.

## Steps

The following steps are executed in this stage:

- [abapEnvironmentCreateSystem](../../../steps/abapEnvironmentCreateSystem.md)

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
```
