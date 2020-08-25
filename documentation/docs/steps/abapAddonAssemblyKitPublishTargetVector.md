# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* werte aus pipeline

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml 

TODO scope dazu fügen, 
The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile looks:

```groovy
abapAddonAssemblyKitPublishTargetVector script: this
```
The config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitPublishTargetVector:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    abapAddonAssemblyKitEndpoint: 'https://myabapAddonAssemblyKitEndpoint.com',
```

### Input via the CommonPipelineEnvironment

TODO
ich glaub targetvectorId
```yaml
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
