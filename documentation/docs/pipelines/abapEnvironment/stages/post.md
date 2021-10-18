# Post

This stage deletes the SAP BTP, ABAP environment system created in the `Prepare System` stage.

## Steps

The following steps are executed in this stage:

- [cloudFoundryDeleteService](../../../steps/cloudFoundryDeleteService.md)

## Stage Parameters

The parameter `confirmDeletion` influences, if a manual confirmation is required between the creation and deletion of the system.

| Value | Explanation |
| --- | --- |
| true | Before the system is deleted, a manual confirmation is requried if the pipeline status is not "SUCCESS". |
| false | The system is deleted without manual confirmation. This is the default. |

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for the `Prepare System` stage.

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
  Post:
    confirmDeletion: true
    cfDeleteServiceKeys: true
```

## Extension

This stage can be extended by creating the file `.pipeline/extensions/Post.groovy`. See [extensibility](../../../extensibility.md)
