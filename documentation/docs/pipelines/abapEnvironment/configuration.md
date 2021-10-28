# Configuration

In general, the SAP BTP, ABAP environment pipeline supports different scenarios. The idea is that only configured stages are executed and the user is able to choose the appropriate stages.
In this section, you can learn how to create a configuration in a (GitHub) repository to run an ABAP environment pipeline used for testing. This specific example will create a pipeline, which executes ATC checks after creating a new ABAP environment system. In the end, the system will be deprovisioned.

You can have a look at different pipeline configurations in our [SAP-samples repository](https://github.com/SAP-samples/abap-platform-ci-cd-samples) or learn more about the configuration in the respective stage or step documentation.

| Stage                    | Steps |
|--------------------------|-------|
| Init                     | -     |
| [Initial Checks](stages/initialChecks.md)           | [abapAddonAssemblyKitCheckPV](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV/), [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs/)|
| [Prepare System](stages/prepareSystem.md)           | [abapEnvironmentCreateSystem](https://sap.github.io/jenkins-library/steps/abapEnvironmentCreateSystem/), [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/)|
| [Clone Repositories](stages/cloneRepositories.md)       | [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/)|
| [ATC](stages/ATC.md)                      | [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/)|
| [Build](stages/build.md)                    | [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/), [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages/), [abapEnvironmentAssemblePackages](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssemblePackages/), [abapAddonAssemblyKitRegisterPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitRegisterPackages/), [abapAddonAssemblyKitReleasePackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReleasePackages/), [abapEnvironmentAssembleConfirm](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssembleConfirm/), [abapAddonAssemblyKitCreateTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCreateTargetVector/), [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)|
| [Integration Tests](stages/integrationTest.md)        | [cloudFoundryCreateService](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateService/)|
| [Confirm](stages/confirm.md)                  | -     |
| [Publish](stages/publish.md)                  | [abapAddonAssemblyKitPublishTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitPublishTargetVector/)|
| [Post](stages/post.md)                     | [cloudFoundryDeleteService](https://sap.github.io/jenkins-library/steps/cloudFoundryDeleteService/)|

## 1. Prerequisites

* Configure your Jenkins Server according to the [documentation](https://sap.github.io/jenkins-library/guidedtour/).
* Create a git repository on a host reachable by the Jenkins server (e.g. GitHub.com). The pipeline will be configured in this repository. Create a GitHub User with read access.
* The entitlements for the ABAP environment system are available in the SAP BTP global account and assigned to the subaccount.
* A Cloud Foundry Organization & Space with the allocated entitlements are available.
* A Cloud Foundry User & Password with the required authorization ("Space Developer") in the Organization and Space are available. User and Password were saved in the Jenkins Credentials Store.

## 2. Jenkinsfile

Create a file named `Jenkinsfile` in your repository with the following content:

```
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

The annotation `@Library('piper-lib-os')` is a reference to the Jenkins Configuration, where you configured the project "Piper" library as a "Global Pipeline Library". If you want to **avoid breaking changes** we advise you to use a specific release of the Piper Library instead of the default master branch. This can be achieved by either adapting the configuration (see [documentation](https://sap.github.io/jenkins-library/infrastructure/customjenkins/#shared-library)) or by specifying the release within the annotaion:

```
@Library('piper-lib-os@v1.53.0') _
```

An Overview of the releases of the project "Piper" library can be found [here](https://github.com/SAP/jenkins-library/releases).

## 3. Configuration for the Communication

The communication to the ABAP system is done using a Communication Arrangement. The Communication Arrangement is created during the pipeline stage `Prepare System` after the system creation via the command `cf create-service-key`. The configuration for the command needs to be stored in a JSON file. Create the file `sap_com_0510.json` in the repository with the following content:

```json
{
  "scenario_id": "SAP_COM_0510",
  "type": "basic"
}
```

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/) for more details.

## 4. Configuration for Cloning the repositories

If you have specified the `Clone Repositories` Stage you can make use of a dedicated configuration file containing the repositories to be pulled and the branches to be switched on. The `repositories` flag makes use of such a configuration file and helps executing a Pull, Clone and Checkout of the Branches of the Repositores. Create the file `repositories.yml` with the following structure containing your repositories including the branches for this stage.

```yml
repositories:
- name: '/DMO/GIT_REPOSITORY'
  branch: 'master'
- name: '/DMO/GIT_REPO'
  branch: 'master'
```

You can later use the `repositories.yml` file for the `repositories` parameter in the `Clone Repositories` stage used in chapter [7. Technical Pipeline Configuration](#7-technical-pipeline-configuration).

## 5. Configuration for ATC

Create a file `atcConfig.yml` to store the configuration for the ATC run. In this file, you can specify which packages or software components shall be checked. Please have a look at the step documentation for more details. Here is an example of the configuration:

```yml
atcobjects:
  softwarecomponent:
    - name: "/DMO/REPO"
```

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/) for more details.

## 6. Technical Pipeline Configuration

Create a file `.pipeline/config.yml` where you store the configuration for the pipeline, e.g. apiEndpoints and credentialIds. The steps make use of the credentials store of the Jenkins server. Here is an example of the configuration file:

```yml
general:
  cfApiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
  cfOrg: 'your-cf-org'
  cfSpace: 'yourSpace'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'abapEnvironmentPipeline'
  cfServiceKeyName: 'jenkins_sap_com_0510'
stages:
  Prepare System:
    cfService: 'abap'
    cfServicePlan: 'standard'
    abapSystemAdminEmail: 'user@example.com'
    abapSystemDescription: 'ABAP Environment Q System'
    abapSystemIsDevelopmentAllowed: false
    abapSystemID: 'H02'
    abapSystemSizeOfPersistence: 4
    abapSystemSizeOfRuntime: 1
    cfServiceKeyConfig: 'sap_com_0510.json'
  Clone Repositories:
    strategy: 'Clone'
    repositories: 'repositories.yml'
  ATC:
    atcConfig: 'atcConfig.yml'
steps:
  cloudFoundryDeleteService:
    cfDeleteServiceKeys: true
```

Some stages may only be executed if a certain condition is met. For example: the stage `Prepare System` will only be executed if it is configured in the stages section. Also, the created system will be deprovisioned in the cleanup routine - although it is necessary to configure the step `cloudFoundryDeleteService` as above.

### Prepare system

The example values for the `Prepare System` stage are a suggestion. Please change them accordingly and don't forget to enter your own email address. Please be aware that creating a SAP BTP, ABAP environment instance may incur costs.

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentCreateSystem/) for more details.

### Clone Repositories

If the `Clone Repositories` stage is configured, you can specify the `strategy` that should be performed on the software components and the branches that you have configured in the `respositories.yml` file in step [4. Configuration for Cloning the repositories](#4-configuration-for-cloning-the-repositories). Per default the strategy will be set to `Pull` if not specified. The following strategies are supported and can be used on the software components and branches:

* `Pull`: If you have specified Pull as the strategy the [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/) step will be used
* `Clone`: If you have specified the Clone strategy the [abapEnvironmentCloneGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentCloneGitRepo/) step will be used
* `CheckoutPull`: This strategy performs a Checkout of Branches with the [abapEnvironmentCheckoutBranch](https://sap.github.io/jenkins-library/steps/abapEnvironmentCheckoutBranch/) step followed by a Pull of the Software Component with the [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/) step

Note that you can use the `repositories.yml` file with the `repositories` parameter consistently for all strategies.

The values for `cfApiEndpoint`,`cfOrg` and `cfSpace` can be found in the respective overview pages in the SAP BTP cockpit. The Cloud Foundry credentials, saved in the Jenkins credentials store with the ID `cfCredentialsId`, must refer to a user with the required authorizations ("Space Developer") for the Cloud Foundry organization and space.

## 7. Create a Jenkins Pipeline

On your Jenkins server click on `New Item` to create a new pipeline. Provide a name and select the type `Pipeline`.
On the creation screen for the pipeline, scroll to the section `Pipeline` and select `Pipeline script from SCM`. Provide the URL (and credentials - if required) of the repository, in which you configured the pipeline. Make sure the `Script Path` points to your Jenkinsfile - if you created the Jenkinsfile according to the documentation above, the default value should be correct.

If you want to configure a build trigger, this can be done in the section of the same name. Here is one example: to run the pipeline every night, you can tick the box "Run periodically". In the visible input field, you can specify a shedule. Click on the questionsmark to read the documentation. The following example will result in the pipeline running every night between 3am and 4am.

```
H H(3-4) * * *
```

### Stage Names

The stage name for the extension is usually the displayed name, e.g. `ATC.groovy` or `Prepare System.groovy`. One exception is the generated `Post` stage. While the displayed name is "Declarative: Post Actions", you can extend this stage using `Post.groovy`.

## Extension

You can extend each stage of this pipeline following the [general extensibility documentation](../../extensibility.md) and the specific [ABAP Environment pipeline extensibility documentation](extensibility.md).
