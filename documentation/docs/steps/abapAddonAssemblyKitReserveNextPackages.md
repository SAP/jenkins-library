# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (e.g. S-User) must be stored in the Jenkins Credential Store
* The step needs information about the Software Component Versions for which packages should be reserved.
* This information is provided via the addonDescriptor in the commonPipelineEnvironment where the fields 'name' and 'version' in the repositories list need to be filled.
* The Software Component Versions must be valid.
* The validation is performed and the required information is written to the CommonPipelineEnvironment if you run prior to this step the step [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs)

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitReserveNextPackages script: this
```

The config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitReserveNextPackages:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    abapAddonAssemblyKitEndpoint: 'https://myabapAddonAssemblyKitEndpoint.com',
```

### Input via the CommonPipelineEnvironment

Mandatory fields:

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
    "version":"1.0.1",
    "versionAAK":"",
    "PackageName":"",
    "PackageType":"",
    "SpLevel":"",
    "PatchLevel":"",
    "PredecessorCommitID":"",
    "Status":"",
    "Namespace":"",
    "SarXMLFilePath":""
  },
  {
    "name":"/DMO/REPO_B",
    "tag":"",
    "branch":"",
    "version":"2.1.1",
    "versionAAK":"",
    "PackageName":"",
    "PackageType":"",
    "SpLevel":"",
    "PatchLevel":"",
    "PredecessorCommitID":"",
    "Status":"",
    "Namespace":"",
    "SarXMLFilePath":""
  }
]}
```
