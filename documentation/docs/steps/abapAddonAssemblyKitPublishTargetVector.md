# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* werte aus pipeline: target vector id 
* scope

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

Mandatory fields: 

```json
{"addonProduct":"",
"addonVersion":"",
"addonVersionAAK":"",
"addonUniqueID":"",
"customerID":"",
"AddonSpsLevel":"",
"AddonPatchLevel":"",
"TargetVectorID":"W7Q00207512600000188",
"repositories":[
  {
    "name":"",
    "tag":"",
    "branch":"",
    "version":"",
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
    "name":"",
    "tag":"",
    "branch":"",
    "version":"",
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
