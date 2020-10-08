# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (e.g. S-User) must be stored in the Jenkins Credential Store
* This step needs the Product Version name and the resolved version(version, spslevel and patchlevel).
* It also needs for each Software Component Version which should be part of the Target Vector, the name and the resolved version(version, splevel and patchlevel) as well as the Delivery Package.
* The Delivery Packages must exist in the package registry (status "P") or already as physical packages (status "L" or "R").
* This information is taken from the addonDescriptor in the commonPipelineEnvironment.
* If you run prior to this step the steps: [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs), [abapAddonAssemblyKitCheckPV](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV) and [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages) you will get the needed information.

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
    "Status":"L",
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
    "Status":"R",
    "Namespace":"",
    "SarXMLFilePath":""
  }
]}
```
