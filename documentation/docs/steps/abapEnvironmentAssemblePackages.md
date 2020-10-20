# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* A SAP Cloud Platform ABAP Environment system is available.
  * This can be created manually on cloud foundry.
  * In a pipeline, you can do this, for example, with the step [cloudFoundryCreateService](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateService/).
* Communication Scenario [“SAP Cloud Platform ABAP Environment - Software Assembly Integration (SAP_COM_0582)“](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/26b8df5435c649aa8ea7b3688ad5bb0a.html) is setup for this system.
  * E.g. a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) are configured.
  * This can be done manually through the respective applications on the SAP Cloud Platform ABAP Environment System,
  * or through creating a service key for the system on cloud foundry with the parameters {“scenario_id”: “SAP_COM_0582", “type”: “basic”}.
  * In a pipeline, you can do this, for example, with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You have following options to provide the ABAP endpoint configuration:
  * The host and credentials the Cloud Platform ABAP Environment system itself. The credentials must be configured for the Communication Scenario SAP_COM_0582.
  * The Cloud Foundry parameters (API endpoint, organization, space), credentials, the service instance for the ABAP service and the service key for the Communication Scenario SAP_COM_0582.
  * Only provide one of those options with the respective credentials. If all values are provided, the direct communication (via host) has priority.
* The step needs information about the packages which should be assembled present in the CommonPipelineEnvironment.
  * For each repository/component version it needs the name of the repository, the version, splevel, patchlevel, namespace, packagename, package type, the status of the package, and optional the predecessor commit id.
  * To upload this information to the CommonPipelineEnvironment run prior to this step the steps:
    * [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs/),
    * [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV/).
  * If one of the package is already in status released, the assembly for this package will not be executed.
* The Software Components for which packages are to be assembled need to be present in the system.
  * This can be done manually through the respective applications on the SAP Cloud Platform ABAP Environment System.
  * In a pipeline, you can do this, for example, with the step [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/).
* The packages to be assembled need to be reserved in AAKaaS and the corresponding information needs to be present in CommonPipelineEnvironment. To do so run step [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages/) prior this step.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentAssemblePackages script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentAssemblePackages:
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
```

Or by authenticating against Cloud Foundry and reading the Service Key details from there:

```yaml
steps:
  abapEnvironmentAssemblePackages:
    abapCredentialsId: 'cfCredentialsId',
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
    cfServiceKeyName: 'myServiceKey',
```

### Input via the CommonPipelineEnvironment

```json
{"addonProduct":"",
"addonVersion":"",
"addonVersionAAK":"",
"addonUniqueID":"",
"customerID":"",
"AddonSpsLevel":"",
"AddonPatchLevel":"",
"TargetVectorID":"",
"repositories":[
  {
    "name":"/DMO/REPO_A",
    "tag":"",
    "branch":"",
    "version":"",
    "versionAAK":"0001",
    "PackageName":"SAPK001001REPOA",
    "PackageType":"CPK",
    "SpLevel":"0000",
    "PatchLevel":"0001",
    "PredecessorCommitID":"cbb834e9e03cde177d2f109a6676901972983fbc",
    "Status":"P",
    "Namespace":"/DMO/",
    "SarXMLFilePath":""
  },
  {
    "name":"/DMO/REPO_B",
    "tag":"",
    "branch":"",
    "version":"",
    "versionAAK":"0002",
    "PackageName":"SAPK002001REPOB",
    "PackageType":"CPK",
    "SpLevel":"0001",
    "PatchLevel":"0001",
    "PredecessorCommitID":"2f7d43923c041a07a76c8adc859c737ad772ef26",
    "Status":"P",
    "Namespace":"/DMO/",
    "SarXMLFilePath":""
  }
]}
```
