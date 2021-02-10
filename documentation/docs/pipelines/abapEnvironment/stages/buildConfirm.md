# Build

This stage is responsible for confirming the delivery of the builded ABAP Add-on for the SAP BTP ABAP Environment. The confirm process of the add-on is done on a Steampunk system (using [SAP_COM_0582](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/26b8df5435c649aa8ea7b3688ad5bb0a.html)). After executing this stage successfully, the add-on delivery is confirmed. For more details, please refer to the [scenario description](../../../scenarios/abapEnvironmentAddons.md)).

## Steps

The following steps are executed in this stage:

- [cloudFoundryCreateServiceKey](../../../steps/cloudFoundryCreateServiceKey.md)
- [abapEnvironmentAssembleConfirm](../../../steps/abapEnvironmentAssembleConfirm.md)

## Stage Parameters

There are no specifc stage parameters.

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

## Configuration Example

### config.yml

```yaml
general:
  addonDescriptorFileName: 'addon.yml'
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'myOrgBld'
  cfSpace: 'mySpaceBld'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'bld_system'
  cfServiceKeyName: 'JENKINS_SAP_COM_0582'
stages:
  Build Confirm:
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
