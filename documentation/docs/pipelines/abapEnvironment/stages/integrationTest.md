# Integration Tests

This stage creates an SAP BTP, ABAP environment (Steampunk) system and installs the add-on product, that was built in the `Build` stage.

## Steps

The following steps are executed in this stage:

- [abapEnvironmentCreateSystem](../../../steps/abapEnvironmentCreateSystem.md)
- [cloudFoundryDeleteService](../../../steps/cloudFoundryDeleteService.md)

## Stage Parameters

The parameter `confirmDeletion` influences, if a manual confirmation is required between the creation and deletion of the system.

| Value | Explanation |
| --- | --- |
| true | Before the system is deleted, a manual confirmation is requried. This is the default. |
| false | The system is deleted without manual confirmation. This is currently not recommended. |

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

## Configuration Example

### config.yml

```yaml
general:
  addonDescriptorFileName: 'addon.yml'
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfCredentialsId: 'cfAuthentification'
stages:
  Integration Tests:
    cfOrg: 'myOrgAti'
    cfSpace: 'mySpaceAti'
    cfServiceInstance: 'ati_system'
    cfService: 'abap'
    cfServicePlan: 'saas_oem'
    abapSystemAdminEmail: 'user@example.com'
    abapSystemDescription: 'Add-on Installation Test System'
    abapSystemIsDevelopmentAllowed: false
    abapSystemID: 'ATI'
    abapSystemSizeOfPersistence: 4
    abapSystemSizeOfRuntime: 1
    includeAddon: true
    confirmDeletion: true
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
