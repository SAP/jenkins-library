# Clone Repositories

This stage creates pulls/clones the specified software components (repositories) to the SAP BTP, ABAP environment system. As a prerequisite, the Communication Arrangement [SAP_COM_0510](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/b04a9ae412894725a2fc539bfb1ca055.html) (SAP BTP, ABAP Environment - Software Component Test Integration) is created using the step `cloudFoundryCreateServiceKey`. With the creation of the Communication Arrangement, a User and Password is created on the SAP BTP, ABAP environment system for the APIs that are used in this stage, as well as in the ATC stage.

## Steps

The following steps can be executed in this stage:

- [cloudFoundryCreateServiceKey](../../../steps/cloudFoundryCreateServiceKey.md)
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
stages:
  Clone Repositories:
    repositories: 'repositories.yml'
    strategy: 'Clone'
```

### repositories.yml

```yaml
repositories:
  - name: '/DMO/SWC'
    branch: 'main'
    commitID: 'cd87a3cac2bc946b7629580e58598c3db56a26f8' #optional
    tag: 'myTag' #optional
```
