# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* A SAP Cloud Platform ABAP Environment system is available. On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario “SAP Cloud Platform ABAP Environment - Software Assembly Integration (SAP_COM_0582)“. This can be done manually through the respective applications on the SAP Cloud Platform ABAP Environment System or through creating a service key for the system on cloud foundry with the parameters {“scenario_id”: “SAP_COM_0582", “type”: “basic”}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You can either provide the ABAP endpoint configuration to directly trigger the assembly on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP Cloud Platform ABAP Environment system in Cloud Foundry that contains all the details of the ABAP endpoint to trigger the assembly.
* The step needs information about the packages which should be assembled. This information is provided via the CommonPipelineEnvironment. For each repository/component version it needs the name of the repository, the version, splevel, patchlevel, namespace, packagename, package type, the status of the package, and optional the predecessor commit id.
* These information will be written to the CommonPipelineEnvironment if you run prior to this step the steps [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/) (!correct link will follow once ready!) and [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/) (!correct link will follow once ready!)
* If one of the package is already in status released, the assembly for this package will not be executed

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
