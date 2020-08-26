# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* werte aus pipeline: fürs product: Please provide product name, version, spslevel and patchlevel"
* für swc: Please provide software component name, version, splevel, patchlevel and packagename
* was muss vorher gelaufen sein: check PV, check SCV, reserve next Packages -> vermutlich sonst auch noch register?

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml 

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitCreateTargetVector script: this
```
The config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitCreateTargetVector:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    abapAddonAssemblyKitEndpoint: 'https://myabapAddonAssemblyKitEndpoint.com',
```

### Input via the CommonPipelineEnvironment

```json
{"addonProduct":"/DMO/myAddonProduct",
"addonVersion":"",
"addonVersionAAK":"0003",
"addonUniqueID":"",
"customerID":"",
"AddonSpsLevel":"0001",
"AddonPatchLevel":"0004",
"TargetVectorID":"",
"repositories":[
  {
    "name":"/DMO/REPO_A",
    "tag":"",
    "branch":"",
    "version":"",
    "versionAAK":"0001",
    "PackageName":"SAPK001001REPOA",
    "PackageType":"",
    "SpLevel":"0000",
    "PatchLevel":"0001",
    "PredecessorCommitID":"",
    "Status":"",
    "Namespace":"",
    "SarXMLFilePath":""
  },
  {
    "name":"/DMO/REPO_B",
    "tag":"",
    "branch":"",
    "version":"",
    "versionAAK":"0002",
    "PackageName":"SAPK002001REPOB",
    "PackageType":"",
    "SpLevel":"0001",
    "PatchLevel":"0001",
    "PredecessorCommitID":"",
    "Status":"",
    "Namespace":"",
    "SarXMLFilePath":""
  }
]}
```
