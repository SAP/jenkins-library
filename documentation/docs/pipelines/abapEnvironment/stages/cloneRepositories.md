# Clone Repositories

This stage creates pulls/clones the specified software components (repositories) to the ABAP Environment system.

!!! caution "Upcoming 2102 release of SAP BTP ABAP Environment"

    With the upcoming 2102 release of SAP BTP ABAP Environment some changes to the backend behavior of the MANAGE_GIT_REPOSITORY service are introduced. Specifically:

      - To pull a software component to a system, the software component needs to be cloned first.
      - It is planned to add the possibility to clone a software component repeatedly with the hotfix collection HFC03 of release 2102

    **Implications for the “abapEnvironmentPipeline”:**

    If you are using the “Prepare System” stage to create a new ABAP Environment system, it is no longer possible to use the “Clone Repositories” stage with the “Pull” strategy or with the default strategy (no strategy specified). Please use the strategy “Clone” instead. For more information, read the stage documentation below.
    The strategy “AddonBuild” will execute the abapEnvironmentCloneGitRepo instead of the previous logic. No configuration changes should be necessary.

    Please be aware that a repeated execution of a pipeline using the strategy “Clone” or “AddonBuild” will not be possible until hotfix collection HFC03 (planned).
    The recommended workaround is to replace the strategy “AddonBuild” with “CheckoutPull”, whenever the system from a previous pipeline run is reused.

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
| AddonBuild | This is recommended, if the stage has to handle both newly create systems and static systems. |

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
