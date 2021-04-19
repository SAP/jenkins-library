# Build and Publish Add-on Products on SAP BTP, ABAP Environment

!!! caution Current limitations
    * Add-on Descriptor File: Please use the long commit ID in the commit ID field in the add-on descriptor file currently, if you are using the short commit ID the build             process will fail.
      To retrieve the long commit id, go into the Manage Software Components app, navigate to the branch, select the commit in the list of commits, field "Long Commit ID"             becomes available.
    * TABU entries are currently not reflected in the object list during add-on build. With ABAP Environment release 2105 HFC 03 it is planned that this is supported.
      Besides that, deletion of TABU is currently not supported via gCTS – this will be supported at a later point in time.

## Introduction

This scenario describes how an add-on for the SAP BTP, ABAP environment is built. It is intended for SAP partners who want to provide a Software as a Service (SaaS) solution on the SAP BTP using the ABAP Environment. Therefore, a partner development contract (see [SAP PartnerEdge Test, Demo & Development Price List](https://partneredge.sap.com/en/library/assets/partnership/sales/order_license/pl_pl_part_price_list.html)) is required. This page aims to provide an overview of the build process of the add-on.

The development on SAP BTP, ABAP environment systems is done within [“software components”](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/58480f43e0b64de782196922bc5f1ca0.html) (also called: “repositories”). The add-ons being built in this scenario are made up by one or multiple software components combined to an add-on product. The “ABAP environment pipeline” can be used to build and publish the add-on product. Please read on for more details about the Add-on Product and the build process.

Of course, this tackles only the upstream part of the SaaS solution lifecycle. Once the add-on is published, it can be consumed as a [multitenant application in ABAP environment](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/195031ff8f484b51af16fe392ec2ae6e.html).

## The Add-on Product

The installation and maintenance of ABAP software is done / controlled via add-on product versions. An **add-on product version** is a „bundle" of software component versions made available at the same time for implementing a well-defined scope of functionality. It is the technical / delivery view on a software portfolio.

### Add-on Product Version

An add-on product version is defined by a name and a version string. The name of an add-on product is a string with a maximum of 30 characters and consists of the [namespace](https://launchpad.support.sap.com/#/notes/84282) and a freely chooseble part - `/NAMESPC/PRODUCTX`. The version string consists of three numbers separated by a dot - `1.2.0`. The numbers in the version string have a hierarchic relationship:

- The first number denotes the release. Release deliveries contain the complete scope of functionality. It is possible to change the software component version bundle in a new release.
- The second number denotes the Support Package Stack level. A Support Package stack consists of Support Package deliveries of the contained software component versions. It is not possible to change the software component version bundle in such a delivery.
- The third number denotes the Patch level. A Patch delivery contains Patch deliveries of the contained software component versions.

!!! note "Development on SAP BTP, ABAP environment"
    As you may know, the development in the SAP BTP, ABAP environment is done within [software component](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/58480f43e0b64de782196922bc5f1ca0.html). A software component is self-contained, and a reduced set of [objects and features of the ABAP programming language](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/c99ba0d28a1a4747b8f47eda06c6b4f1.html) can be used.
    The software component and development objects must be created in a namespace, so that clashes between software of different vendors and SAP are avoided. Therefore, a namespace must be reserved before the development can start. [SAP note 105132](https://launchpad.support.sap.com/#/notes/105132) describes the namespace reservation process. The namespace must be reserved for the same customer number under which the “SAP BTP, ABAP ENVIRONMENT” tenants are licensed.

### Software Component Version

A **software component version** is a technically distinguishable unit of software and is installed and patched as a whole. It consists of ABAP development packages and contained objects. Software component versions are delivered via delivery packages. But software component versions are not individual shipment entities. They can only be delivered to customers as part of an [add-on product version](#add-on-product-version).
A software component version is defined by a name and a version string. The name of a software component is string with a maximum of 30 characters and consists of the [namespace](https://launchpad.support.sap.com/#/notes/84282) and a freely chooseble part - `/NAMESPC/COMPONENTA`. The version consists of three numbers separated by a dot - 1.2.0. The numbers in the version string have a hierarchic relationship:

- The first number denotes the release. Release deliveries contains the whole software component and deliver new and enhancements of existing functionalities. They are delivered with delivery packages of type [“Installation Package”](https://help.sap.com/viewer/9043aa5d2f834ad385e1cdfdadc06b6f/5.0.4.7/en-US/6082f55473568c77e10000000a174cb4.html).
- The second number denotes the Support Package level. Support Package deliveries contain a larger collection of corrections and may contains smaller functional enhancements. They are delivered with delivery packages of type [“Component Support Package”](https://help.sap.com/viewer/9043aa5d2f834ad385e1cdfdadc06b6f/5.0.4.7/en-US/6082f55473568c77e10000000a174cb4.html).
- The third number denotes the Patch level. Patch deliveries shall only contain small corrections. They are shipped with delivery packages of type “Correction Package”.

**Note:** The needed type of delivery does not need to be chosen manually; it is automatically determined by the delivery tools.

### Target Vector

As explained above, the shipment of a software takes place via add-on product versions. The delivered content of an add-on product version is defined in a target vector, which is used by the deployment tools. The target vector is derived from the addon.yml (more on that below) and contains the following information:

- Product name
- Product release
- Product Support Package stack and Patch level
- A list of contained software component versions with
  - Software component name
  - Software component release
  - Delivery Package, which delivers the versions

## Building the Add-on Product

The build process of an add-on product is orchestrated by a Jenkins Pipeline, the “ABAP environment pipeline” provided in this project. To run this pipeline, it only needs to be configured – which will be explained in the sections “Prerequisites” and “Configuration”.

![ABAP Environment Pipeline Build](../images/abapEnvironmentBuildPipeline.png)

The pipeline consists of different steps responsible for a single task. The steps themselves are grouped thematically into different stages. For example, early in the pipeline, the ABAP environment system needs to be created and the communication needs to be set up. This is done in the “Prepare System” stage. You can read more about the different stages in the ABAP environment pipeline [documentation](https://sap.github.io/jenkins-library/pipelines/abapEnvironment/introduction/).

Different services and systems are required for the add-on build process.

### Delivery Tools

With the following tools the add-on deliveries are created.

#### Assembly System

First the ABAP system responsible for the add-on assembly. It is created during the pipeline and deleted in the end. All actions related to the ABAP source code are executed on this system, e.g. running checks with the ABAP test cockpit (ATC) or the physical build of the software components. There are two communication scenarios containing the different APIs of the ABAP environment system: [Test Integration](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/b04a9ae412894725a2fc539bfb1ca055.html) and [Software Assembly Integration](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/b04a9ae412894725a2fc539bfb1ca055.html).
The assembly system should be of [service type abap](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/f0163565eb554f009f990652ca41d1c6.html) and be provisioned with parameter `is_development_allowed = false` to prevent local changes.

#### Add-on Assembly Kit as a Service (=AAKaaS)

The Add-on Assembly Kit as a Service is responsible for registering and publishing the add-on product. It is accessible via APIs with an S-User.

### Deployment Tools

With these SAP tools the assembled add-on deliveries are deployed to ABAP systems, for example into the [installation test system](#installation-test-system).

#### Installation Test System

In order to verify that the delivery packages included in the add-on product version being built are installable, a target vector is published in "test" scope. In the *Integration Tests* stage an ABAP system of service plan `saas_oem` is created. This ABAP system makes it possible to install a specific add-on product version into an ABAP system that is provisioned. The installation test system should be be provisioned with parameter `is_development_allowed = false` to prevent local changes.

### Prerequisites

There are several parts that are required to run the pipeline for building an ABAP Environment Add-on.

#### Jenkins Server

The pipeline responsible for building ABAP add-ons has been created specifically for [Jenkins](https://www.jenkins.io). Therefore, a Jenkins Server is required. The [piper project](https://sap.github.io/jenkins-library/guidedtour/) provides a Jenkins image, which already includes the necessary configurations. Of course, it is also possible to [configure an existing server](https://sap.github.io/jenkins-library/infrastructure/customjenkins/).

#### Git Repository

The pipeline configuration is done in a git repository (for example on GitHub). This repository needs to be accessed by the Jenkins Server. If the repository is password protected, the user and password (or access token) should be stored in the Jenkins Credentials Store (Manage Jenkins &rightarrow; Manage Credentials).

#### Add-on Assembly Kit as a Service (=AAKaaS)

The communication with the AAKaaS needs a technical communication user. The creation and activation of such a user is described in [SAP note 2174416](https://launchpad.support.sap.com/#/notes/2174416). Make sure that this technical communication user is assigned to the customer number under which the “SAP CP ABAP ENVIRONMENT” tenants are licensed and for which the development namespace was reserved. The user and password need to be stored in the Jenkins Credentials Store.

#### Cloud Foundry Access

ABAP environment systems are created in the SAP BTP cockpit. For this pipeline, the creation and deletion of the systems are automated via the Cloud Foundry command line interface: [cf CLI](https://docs.cloudfoundry.org/cf-cli/). For this to work, two things need to be configured:

- Cloud Foundry needs to be enabled on subaccount level. This can be done on the Subaccount Overview page. The subaccount is then mapped to a “Cloud Foundry Organization”, for which you must provide a suitable name during the creation. Have a look at the [documentation](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/dc18bac42270468d84b6c030a668e003.html) for more details.
- A (technical) user is required to access the SAP BTP via the cf CLI. The user needs to be a [member of the global account](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/4a0491330a164f5a873fa630c7f45f06.html) and has to have the [Space Developer](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/967fc4e2b1314cf7afc7d7043b53e566.html) role. The user and password need to be stored in the Jenkins Credentials Store.

Later, during the pipeline configuration, you will specify the Service Plan, which will be used for the creation of an ABAP environment system. Please make sure, that there are enough entitlements for this [Service Plan in the Subaccount](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/c40cb18aeaa343389036fdcdd03c41d0.html).

#### Register Add-on Product for a Global Account

The add-on product needs to be registered with SAP in order to be installable in the desired global account. More details will follow soon.

#### Project "Piper" Library Version to SAP BTP, ABAP Environment Dependency

SAP BTP ABAP environment releases might require certain versions of the project "Piper" Library. More Information can be found in [SAP Note 3032800](https://launchpad.support.sap.com/#/notes/3032800).

### Configuration

In the following subsections, the pipeline configuration for this scenario is explained. To get a general overview on the ABAP environment pipeline configuration, have a look [here](https://sap.github.io/jenkins-library/pipelines/abapEnvironment/configuration/). In addition to the following sections explaining the configuration, there will be an example repository including all required files.

#### Jenkinsfile

This file is the entry point of the pipeline. It should look like this:

```Groovy
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

The first line defines that the shared library, named “piper-lib-os” in the Jenkins Configuration, will be used. This is a reference to the [/SAP/Jenkins-library](https://github.com/SAP/jenkins-library/) of project "Piper".

!!! note "Jenkins Library Version"
    If desired, a specific release of this library can be requested: e.g. release 1.93.0 with `@Library('piper-lib-os@v1.93.0') _`. As the library is an Open Source project, it is possible that incompatible changes are introduced. If you want to avoid this, it is recommended to use such a specific release. If no release is specified, the newest version of the Jenkins-library will be used (pulled from the master branch).

The second line `abapEnvironmentPipeline script: this` defines that the predefined “ABAP environment pipeline” will be executed.

#### Pipeline configuration file

A configuration file `.pipeline/config.yml` is used to provide all required values to run the pipeline. This includes - for example - different endpoints or credential IDs of user and password values stored in the [Jenkins Credentials Store](https://www.jenkins.io/doc/book/using/using-credentials/). If a complex configuration is necessary, a separate configuration file is required, which will also be referenced in the `config.yml` file.

#### Add-on descriptor file

The build process is controlled by an add-on descriptor file called `addon.yml`. This file must be created manually and must be stored in the GIT repository of the pipeline. It must contain information about the to-be-delivered [add-on product version](#add-on-product-version) and the contained [software component versions](#software-component-version). Below, you see an example:

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

Explanation of the keys:

- `addonProduct`: this is the technical name of the add-on product
- `addonVersion`: This is the technical version of the add-on product `<product version>.<support package stack level>.<patch level>`

The section “repositories” contains one or multiple software component versions:

- `name`: the technical name of the software component
- `branch`: this is the release branch from the git repository
- `version`: this is the technical software component version `<software component version>.<support package level>.<patch level>`

##### Versioning Rules

For the development and the provisioning of product-/software component versions, it is necessary to ensure, that there are no gaps within the version and level counters. Therefore, only a continuous increase in version numbers is allowed. The following examples show valid and invalid cases, respectively:

Valid increase:

- 1.0.0 to 2.0.0
- 1.1.2 to 2.0.0
- 2.0.0 to 2.0.1
- 2.1.0 to 2.2.0
- 2.1.1 to 2.1.2

Invalid increase:

- 1.0.0 to 3.0.0 (version 2.0.0 is missing; therefore, a product/component version is missing)
- 1.1.2 to 2.1.0 (version 2.0.0 is missing; therefore, a product/component version is missing)
- 2.0.0 to 2.0.2 (version 2.0.1 is missing; therefore, a patch level is missing)
- 2.1.0 to 2.3.0 (version 2.2.0 is missing; therefore, a support package level is missing)
- 2.1.1 to 2.1.3 (version 2.1.2 is missing; therefore, a patch level is missing)

#### Jenkins Job

Once the configuration in the git repository is completed, the pipeline on the Jenkins Server can be created. On your Jenkins server click on “New Item” to create a new pipeline. Provide a name and select the type “Pipeline”. On the creation screen for the pipeline, scroll to the section Pipeline and select “Pipeline script from SCM”. Provide the URL (and credentials - if required) of the repository in which you configured the pipeline. Make sure the “Script Path” points to your Jenkinsfile - if you created the Jenkinsfile according to the documentation above, the default value should be correct.
Make sure to check the general option "Do not allow concurrent builds" in order to prevent concurrent add-on build processes for the same version.

### Example

Please have a look at the configuration example in this [GitHub repository](https://github.com/SAP-samples/abap-platform-ci-cd-samples/tree/addon-build).

## Troubleshooting

If you encounter an issue with the pipeline itself, please open an issue in [GitHub](https://github.com/SAP/jenkins-library/issues).
If the pipelines receives the error from a backend system, please open a [support incident](https://launchpad.support.sap.com/#/notes/1296527) on the respective component:

| Stage                    | Steps | Support Component |
|--------------------------|-------|-------------------|
| Initial Checks           | [abapAddonAssemblyKitCheckPV](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV/), [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs/)| BC-UPG-OCS |
| Prepare System           | [abapEnvironmentCreateSystem](https://sap.github.io/jenkins-library/steps/abapEnvironmentCreateSystem/), [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/)| BC-CP-ABA |
| Clone Repositories       | [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/)| BC-CP-ABA-SC |
| ATC                      | [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/)| BC-DWB-TOO-ATF |
| Build                    | [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/)| BC-CP-ABA |
|                          | [abapEnvironmentAssemblePackages](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssemblePackages/)| BC-UPG-ADDON |
|                          | [abapAddonAssemblyKitReleasePackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReleasePackages/), [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages/), [abapAddonAssemblyKitRegisterPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitRegisterPackages/), [abapAddonAssemblyKitCreateTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCreateTargetVector/), [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)| BC-UPG-OCS |
|                          | [abapEnvironmentAssembleConfirm](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssembleConfirm/)| BC-UPG-ADDON |
| Integration Tests        | [abapEnvironmentCreateSystem](https://sap.github.io/jenkins-library/steps/abapEnvironmentCreateSystem/), [cloudFoundryDeleteService](https://sap.github.io/jenkins-library/steps/cloudFoundryDeleteService/)| BC-CP-ABA |
| Publish                  | [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)| BC-UPG-OCS |
| Post                     | [cloudFoundryDeleteService](https://sap.github.io/jenkins-library/steps/cloudFoundryDeleteService/)| BC-CP-ABA |
