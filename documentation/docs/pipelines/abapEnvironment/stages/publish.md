# Publish

This stage publishes an add-on for the ABAP Environment and confirms the delivery (please refer to the [scenario description](../../../scenarios/abapEnvironmentAddons.md)).

## Steps

The following steps are executed in this stage:

- [abapAddonAssemblyKitPublishTargetVector](../../../steps/abapAddonAssemblyKitPublishTargetVector.md)
- [abapEnvironmentAssembleConfirm](../../../steps/abapEnvironmentAssembleConfirm.md)

## Stage Parameters

There are no specifc stage parameters.

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

## Configuration Example

### config.yml

```yaml
general:
  abapAddonAssemblyKitCredentialsId: 'TechUserAAKaaS'
  addonDescriptorFileName: 'addon.yml'
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'myOrgBld'
  cfSpace: 'mySpaceBld'
  cfCredentialsId: 'cfAuthentification'  
stages:
  Publish:
    targetVectorScope: 'P'
    cfServiceKeyName: 'JENKINS_SAP_COM_0582'
    cfServiceKeyConfig: 'sap_com_0582.json'
```

### addon.yml

```yaml
---
addonProduct: /DMO/PRODUCT1
addonVersion: 1.0.0
repositories:
  - name: /DMO/SWC
    branch: v1.0.0
    version: 1.0.0
```

### sap_com_0582.json

```json
{
  "scenario_id": "SAP_COM_0582",
  "type": "basic"
}
```