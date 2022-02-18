# Continuous Testing on SAP BTP, ABAP Environment

## Introduction

This scenario describes how to test ABAP development for the SAP BTP, ABAP environment (also known as Steampunk). In Steampunk, the development is done within [“software components”](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/58480f43e0b64de782196922bc5f1ca0.html) (also called: “repositories”) and "transported" via git-based approaches. The [ABAP environment pipeline](../pipelines/abapEnvironment/introduction.md) is a predefined pipeline, which can be used to import ABAP development into a quality system and execute tests.

## Pipeline

For this scenario three stages of the ABAP environment pipeline are relevant: "Prepare System", "Clone Repositories" and "ATC".

### Prepare System

The pipeline starts with the stage "Prepare System". This stage, however, is optional.  **If this stage is active**, a new Steampunk system is created for each pipeline execution. This has the advantage that each test runs on a fresh system without a history. On the other hand, the duration of each pipeline execution will increase as the system provisioning takes a significant amount of time. **If this stage is not active**, you have to provide a prepared Steampunk (quality) system for the other stages. Then, each pipeline execution runs on the same system. Of course, the system has a history, but the pipeline duration will be shorter. Please also consider: the total costs may increase for a static system in contrast to a system, which is only active during the pipeline.

### Clone Repositories

This stage is responsible for cloning (or pulling) the defined software components (repositories) to the system.

### Run Tests

This stage will trigger the execution of the `ATC` and `AUnit` stages in parallel. Please find more information on the respective stages below.

### ATC

In this stage, ATC checks can be executed using [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/). The step can receive software components or packages.

The results are returned in the checkstyle format and can be displayed using the [Warnings Next Generation Plugin](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#warnings-next-generation-plugin). To display the results it is necessary to [extend the ATC stage via the Checkstyle/Warnings Next Generation Plugin](https://www.project-piper.io/pipelines/abapEnvironment/extensibility/#1-extend-the-atc-stage-via-the-checkstylewarnings-next-generation-plugin).

### AUnit

This stage will triger an AUnit run on the SAP BTP, APAB environment system. You can configure the object set that should be checked during the AUnit run. The results of the test run are returned in the "JUnit" format. It is possible to further visualize the AUnit test run results with the help of a stage extension.

## Prerequisites

There are several parts that are required to run the pipeline.

### Jenkins Server

The pipeline for testing software components has been created specifically for [Jenkins](https://www.jenkins.io). Therefore, a Jenkins server is required. The [project "Piper"](https://sap.github.io/jenkins-library/guidedtour/) provides a Jenkins image, which already includes the necessary configurations. Of course, it is also possible to [configure an existing server](https://sap.github.io/jenkins-library/infrastructure/customjenkins/).

### Git Repository

The pipeline configuration is done in a git repository (for example on GitHub). This repository needs to be accessed by the Jenkins server. If the repository is password protected, the user and password (or access token) should be stored in the Jenkins Credentials Store (Manage Jenkins &rightarrow; Manage Credentials).

### Cloud Foundry Access

ABAP environment systems are created in the SAP BTP cockpit. For this pipeline, the creation and deletion of the systems are automated via the Cloud Foundry Command Line Interface: [cf CLI](https://docs.cloudfoundry.org/cf-cli/). For this to work, two things need to be configured:

- Cloud Foundry needs to be enabled on subaccount level. This can be done on the Subaccount Overview page. The subaccount is then mapped to a “Cloud Foundry Organization”, for which you must provide a suitable name during the creation. Have a look at the [documentation](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/dc18bac42270468d84b6c030a668e003.html) for more details.
- A (technical) user is required to access the SAP BTP via the cf CLI. The user needs to be a [member of the global account](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/4a0491330a164f5a873fa630c7f45f06.html) and has to have the [Space Developer](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/967fc4e2b1314cf7afc7d7043b53e566.html) role. The user and password need to be stored in the Jenkins Credentials Store.

During the pipeline configuration, you will specify the service plan, which will be used for the creation of an ABAP environment system. Please make sure, that there are enough entitlements for this [Service Plan in the Subaccount](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/c40cb18aeaa343389036fdcdd03c41d0.html).

## Configuration

Please refer to the [configuration page](../pipelines/abapEnvironment/configuration.md).

## Example

Please have a look at the configuration examples to run ATC checks on a [transient system](https://github.com/SAP-samples/abap-platform-ci-cd-samples/tree/atc-transient) or on a [permanent system](https://github.com/SAP-samples/abap-platform-ci-cd-samples/tree/atc-static).
