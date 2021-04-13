# Publish

This stage publishes an add-on for the SAP BTP, ABAP environment (please refer to the [scenario description](../../../scenarios/abapEnvironmentAddons.md)).

## Steps

The following steps are executed in this stage:

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
stages:
  Publish:
    targetVectorScope: 'P'
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
