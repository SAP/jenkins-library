# Publish

This stage publishes an add-on for the ABAP Environment (please refer to the [scenario description](../../../scenarios/abapEnvironmentAddons.md)).

## Steps

The following steps are executed in this stage:

- [abapAddonAssemblyKitPublishTargetVector](../../steps/abapAddonAssemblyKitPublishTargetVector.md)

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

```yaml
---
addonProduct: /DMO/PRODUCT1
addonVersion: 1.0.0
repositories:
   - name: /DMO/SWC
     branch: v1.0.0
     version: 1.0.0
```
