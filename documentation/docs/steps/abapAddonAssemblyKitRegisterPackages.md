# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (e.g. S-User) must be stored in the Jenkins Credential Store
* This step needs the names of the packages which should be registered. For each package a SAR archive with the data file and metadata XML must be provided.
* The package names and their status are taken from the addonDescriptor in the commonPipelineEnvironment, as well as the SarXMLFilePath with the path to the SAR file.
* The information will be written to the commonPipelineEnvironment if you run prior to this step the step [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages)
* The SAR archive is produced if you run the step [abapEnvironmentAssemblePackages](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssemblePackages)

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitRegisterPackages script: this
```

The config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitRegisterPackages:
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
    "name":,
    "tag":"",
    "branch":"",
    "version":"",
    "versionAAK":"",
    "PackageName":"SAPK001001REPOA",
    "PackageType":"",
    "SpLevel":"",
    "PatchLevel":"",
    "PredecessorCommitID":"",
    "Status":"P",
    "Namespace":"",
    "SarXMLFilePath":".pipeline/commonPipelineEnvironment/SAPK001001REPOA.SAR"
  },
  {
    "name":"",
    "tag":"",
    "branch":"",
    "version":"",
    "versionAAK":"",
    "PackageName":"SAPK002001REPOB",
    "PackageType":"",
    "SpLevel":"",
    "PatchLevel":"",
    "PredecessorCommitID":"",
    "Status":"R",
    "Namespace":"",
    "SarXMLFilePath":".pipeline/commonPipelineEnvironment/SAPK002001REPOB.SAR"
  }
]}
```
