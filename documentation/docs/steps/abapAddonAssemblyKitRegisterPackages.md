# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* werte aus der pipeline: package name, sar file und path to sarfile
* schritte vorher: Reserve next packages, assembly

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
    "Status":"",
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
    "Status":"",
    "Namespace":"",
    "SarXMLFilePath":".pipeline/commonPipelineEnvironment/SAPK002001REPOB.SAR"
  }
]}
```
