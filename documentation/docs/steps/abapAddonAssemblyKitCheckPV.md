# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* addon.yml

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml 

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitCheckPV script: this
```
The config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitCheckPV:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    abapAddonAssemblyKitEndpoint: 'https://myabapAddonAssemblyKitEndpoint.com',
    addonDescriptorFileName: '.pipeline/addon.yml'
```

### Input via the CommonPipelineEnvironment

TODO
was muss da sein? je nachdem ob cv davor oder danach ist
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
