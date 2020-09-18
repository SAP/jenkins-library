# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (e.g. S-User) must be stored in the Jenkins Credential Store
* This step needs an existing Target Vector as well as the scope where it should be published.
* The Target Vector ID is taken from the addonDescriptor in the commonPipelineEnvironment.
* If you run prior to this step the step [abapAddonAssemblyKitCreateTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCreateTargetVector), the Target Vector will be created and its ID will be written to the commonPipelineEnvironment

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile looks:

```groovy
abapAddonAssemblyKitPublishTargetVector(
                    targetVectorScope: 'T',
                    script: this,
                    )
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
