# ABAP Environment Pipeline

![ABAP Environment Pipeline](../../images/abapPipelineOverview.png)

The goal of the ABAP Environment Pipeline is to enable Continuous Integration for the SAP Cloud Platform ABAP Environment, also known as Steampunk.
The pipeline contains several stages and supports different scenarios. The general idea is that the user can choose a subset of these stages, which fits her/his use case, for example running nightly ATC checks or building an ABAP Add-on for Steampunk.

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
| Initial Checks           | [abapAddonAssemblyKitCheckPV](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV/), [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs/)|
| Prepare System           | [abapEnvironmentCreateSystem](https://sap.github.io/jenkins-library/steps/abapEnvironmentCreateSystem/), [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/)|
| Clone Repositories       | [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/)|
| ATC                      | [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/)|
| Build                    | [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/), [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages/), [abapEnvironmentAssemblePackages](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssemblePackages/), [abapAddonAssemblyKitRegisterPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitRegisterPackages/), [abapAddonAssemblyKitReleasePackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReleasePackages/), [abapAddonAssemblyKitCreateTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCreateTargetVector/), [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)|
| Integration Tests        | [cloudFoundryCreateService](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateService/)|
| Confirm                  | -     |
| Publish                  | [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)|
| Post                     | [cloudFoundryDeleteService](https://sap.github.io/jenkins-library/steps/cloudFoundryDeleteService/)|

Below you can find more details about the different stages. [Here](configuration.md) you can find more information about how to configure your pipeline.

### Init

In this stage, the pipeline is initialized. Nothing to see here.

### Initial Checks

This stage is executed, if the "Build" stage is configured. It contains checks to verify the validity of the provided Add-on Descriptor.

### Prepare System

In this stage, the ABAP Environment system is created. This is done with the abapEnvironmentCreateSystem step. After the system creation, the Communication Arrangement [SAP_COM_0510](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/b04a9ae412894725a2fc539bfb1ca055.html) (SAP Cloud Platform ABAP Environment - Software Component Test Integration) is created using the step cloudFoundryCreateServiceKey. With the creation of the Communication Arrangement, a User and Password is created on the ABAP Environment system for the APIs that are used in the following stages.

### Clone Repositories

As a default we assume that the ABAP Environment system is already configured and all software components are cloned and the latest change of the respective software components should be pulled with the abapEnvironmentPullGitRepo step.
In this stage, the software components / Git repositories are then pulled to the ABAP Environment system using the step abapEnvironmentPullGitRepo.
The step can receive a list of software components / repositories and pulls them successively.
If the software components have not been cloned on the ABAP Environment system yet or you want to e.g. checkout a different Branch you can make use of the `strategy` stage parameter and perform other steps and step orders.
Please refer to the Configuration section for the abapEnvironment Pipeline or the respective documentations for the [abapEnvironmentCheckoutBranch](https://sap.github.io/jenkins-library/steps/abapEnvironmentCheckoutBranch/), [abapEnvironmentCloneGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentCloneGitRepo/) and [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/) steps.

Either way, if you chose a dedicated strategy or the default Pull variant you can optionally provide a dedicated configuration file, e.g. `repositories.yml`, containing the repositories to be cloned and the branches to be switched to. This file can be used consistently for all strategies.

### ATC

In this stage, ATC checks can be executed using abapEnvironmentRunATCCheck. The step can receive software components or packages (configured in YML file - as described in [configuration](configuration.md)). The results are returned in the checkstyle format. With the use of a pipeline extension, quality gates can be configured (see [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/) or the "Extensions" section in the [configuration](configuration.md)).

### Build

This stage is responsible for building an ABAP Add-on for the SAP Cloud Platform ABAP Environment. The build process of the add-on is done on a Steampunk system (using [SAP_COM_0582](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/26b8df5435c649aa8ea7b3688ad5bb0a.html)) with the help of the ABAP Add-on Assembly Kit as a Service (AAKaaS). After executing this stage successfully, the add-on is ready to be tested.

### Integration Tests

This stage is intended to be used for testing the add-on built in the "Build" stage. Nevertheless, it can be configured seperately. In this stage, another ABAP Environment system is created including the add-on (if configured correctly).

### Confirm

This stage is executed if the stage "Publish" is configured. In this stage a manual confirmation is prompted to confirm the publishing of the add-on.

### Publish

In this stage the add-on built with this pipeline is published. After that, it is ready to be delivered to productive systems.

### Post

At the end of every pipeline (successful or unsuccessful), the system is deprovisioned using the step cloudFoundryDeleteService.
