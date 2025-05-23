# Build

This stage is responsible for building an ABAP add-on for the SAP BTP, ABAP environment. The build process of the add-on is done on a Steampunk system (using [SAP_COM_0582](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/26b8df5435c649aa8ea7b3688ad5bb0a.html)) with the help of the SAP Add-On Assembly Kit as a Service (AAKaaS). After executing this stage successfully, the add-on is ready to be tested. For more details, please refer to the [scenario description](../../../scenarios/abapEnvironmentAddons.md)).

## Steps

The following steps are executed in this stage:

- [cloudFoundryCreateServiceKey](../../../steps/cloudFoundryCreateServiceKey.md)
- [abapEnvironmentAssemblePackages](../../../steps/abapEnvironmentAssemblePackages.md)
- [abapAddonAssemblyKitRegisterPackages](../../../steps/abapAddonAssemblyKitRegisterPackages.md)
- [abapAddonAssemblyKitReleasePackages](../../../steps/abapAddonAssemblyKitReleasePackages.md)
- [abapEnvironmentAssembleConfirm](../../../steps/abapEnvironmentAssembleConfirm.md)
- [abapAddonAssemblyKitCreateTargetVector](../../../steps/abapAddonAssemblyKitCreateTargetVector.md)
- [abapAddonAssemblyKitPublishTargetVector](../../../steps/abapAddonAssemblyKitPublishTargetVector.md)

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
  cfServiceInstance: 'bld_system'
  cfServiceKeyName: 'JENKINS_SAP_COM_0948'
stages:
  Build:
    cfServiceKeyName: 'JENKINS_SAP_COM_0582'
    cfServiceKeyConfig: 'sap_com_0582.json'
```

### addon.yml

```YAML
---
addonProduct: /NAMESPC/PRODUCTX
addonVersion: 1.2.0
repositories:
  - name: /NAMESPC/COMPONENTA
    branch: v1.2.0
    version: 1.2.0
    commitID: 7d4516e9
  - name: /NAMESPC/COMPONENTB
    branch: v2.0.0
    version: 2.0.0
    commitID: 9f102ffb
```

### sap_com_0582.json

```json
{
  "scenario_id": "SAP_COM_0582",
  "type": "basic"
}
```
