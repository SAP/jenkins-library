# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* werte aus der pipeline: package name
* steps vorher -> unklar, package name krieg ich aus reserve next, aber muss es zb eventuell im status L sein um released zu werden? dann müsste register gelaufen sein, eigentlich muss natürlich auch die assembly gelaufen sein

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml 

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitReleasePackages script: this
```
The config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitReleasePackages:
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
    "name":"",
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
    "SarXMLFilePath":""
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
    "SarXMLFilePath":""
  }
]}
```
