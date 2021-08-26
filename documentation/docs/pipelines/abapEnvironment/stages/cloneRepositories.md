# Clone Repositories

This stage creates pulls/clones the specified software components (repositories) to the SAP BTP, ABAP environment system.

## Steps

The following steps can be executed in this stage:

- [abapEnvironmentPullGitRepo](../../../steps/abapEnvironmentPullGitRepo.md)
- [abapEnvironmentCloneGitRepo](../../../steps/abapEnvironmentCloneGitRepo.md)
- [abapEnvironmentCheckoutBranch](../../../steps/abapEnvironmentCheckoutBranch.md)

## Stage Parameters

The parameter `strategy` influences, which steps will be executed. Possible values are:

| Value | Explanation |
| --- | --- |
| Clone | The step `abapEnvironmentCloneGitRepo` will be executed. This is recommended, if a new system was created in the `Prepare System` stage. |
| Pull | The step `abapEnvironmentPullGitRepo` will be executed. This is recommended, if a static system is used. The software component should be cloned beforehand. |
| CheckoutPull | The step `abapEnvironmentCheckoutBranch`, followed by `abapEnvironmentPullGitRepo`, will be executed. The software component should be cloned beforehand. This can be used if the branch may change between pipeline executions. |

## Stage Activation

This stage will be active, if the stage configuration in the `config.yml` contains entries for this stage.

## Configuration Example

It is recommended to use a yml configuration to define the software components. This yml file works for all strategies. If you are building an ABAP add-on, the addon descriptor `addon.yml` can be reused.

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
  Clone Repositories:
    repositories: 'repositories.yml'
    strategy: 'Clone'
```

### repositories.yml

```yaml
repositories:
  - name: '/DMO/SWC'
    branch: 'master'
```
