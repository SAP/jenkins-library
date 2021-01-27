# ABAP Environment Pipeline

![ABAP Environment Pipeline](../../images/abapPipelineOverview.png)

The goal of the ABAP Environment Pipeline is to enable Continuous Integration for the SAP BTP ABAP Environment, also known as Steampunk.
The pipeline contains several stages and supports different scenarios. The general idea is that the user can choose a subset of these stages, which fits her/his use case, for example running nightly ATC checks or building an ABAP Add-on for Steampunk.

!!! caution "Upcoming 2102 release of SAP BTP ABAP Environment"
    With the upcoming 2102 release of SAP BTP ABAP Environment some changes to the backend behavior of the MANAGE_GIT_REPOSITORY service are introduced. Specifically:
    To pull a software component to a system, the software component needs to be cloned first.
    It is planned to add the possibility to clone a software component repeatedly with the hotfix collection HFC03 of release 2102
    **Implications for the “abapEnvironmentPipeline”:**
    If you are using the “Prepare System” stage to create a new ABAP Environment system, it is no longer possible to use the “Clone Repositories” stage with the “Pull” strategy or with the default strategy (no strategy specified). Please use the strategy “Clone” instead. For more information, have a look at the [stage documentation](./stages/cloneRepositories.md).
    The strategy “AddonBuild” will execute the abapEnvironmentCloneGitRepo instead of the previous logic. No configuration changes should be necessary.
    Please be aware that a repeated execution of a pipeline using the strategy “Clone” or “AddonBuild” will not be possible until hotfix collection HFC03 (planned).
    The recommended workaround is to replace the strategy “AddonBuild” with “CheckoutPull”, whenever the system from a previous pipeline run is reused.

## Scenarios

The following scenarios are available.

### Continuous Testing

This scenario is intended to be used improve the software quality through continuous checks and testing. Please refer to the [scenario documentation](../../scenarios/abapEnvironmentTest.md) for more information.

### Building ABAP Add-ons for Steampunk

This scenario is intended for SAP Partners, who want to offer a Software as a Service (SaaS) solution on Steampunk. This is currently the only use case for building ABAP Add-ons and, more specifically, the stages "Initial Checks", "Build", "Integration Tests", "Confirm" and "Publish". Please refer to the [scenario documentation](../../scenarios/abapEnvironmentAddons.md) for more information.

## Pipeline Stages

The following stages and steps are part of the pipeline:

| Stage                    | Steps |
|--------------------------|-------|
| Init                     | -     |
| [Initial Checks](stages/initialChecks.md)           | [abapAddonAssemblyKitCheckPV](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV/), [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs/)|
| [Prepare System](stages/prepareSystem.md)           | [abapEnvironmentCreateSystem](https://sap.github.io/jenkins-library/steps/abapEnvironmentCreateSystem/), [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/)|
| [Clone Repositories](stages/cloneRepositories.md)       | [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/)|
| [ATC](stages/ATC.md)                      | [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/)|
| [Build](stages/build.md)                    | [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/), [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages/), [abapEnvironmentAssemblePackages](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssemblePackages/), [abapAddonAssemblyKitRegisterPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitRegisterPackages/), [abapAddonAssemblyKitReleasePackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReleasePackages/), [abapAddonAssemblyKitCreateTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCreateTargetVector/), [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)|
| [Integration Tests](stages/integrationTest.md)        | [cloudFoundryCreateService](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateService/)|
| [Confirm](stages/confirm.md)                  | -     |
| [Publish](stages/publish.md)                  | [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)|
| [Post](stages/post.md)                     | [cloudFoundryDeleteService](https://sap.github.io/jenkins-library/steps/cloudFoundryDeleteService/)|

Please navigate to a stage or step to learn more details. [Here](configuration.md) you can find a step-by-step example on how to configure your pipeline.
