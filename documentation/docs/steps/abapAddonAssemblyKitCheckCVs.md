# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

A detailed description of all prerequisites of the scenario including those needed for this step can be found in the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).

* The credentials to access the AAKaaS (e.g. S-User) must be stored in the Jenkins Credential Store
* The step needs an addon.yml containing information about the Product Version and corresponding Software Component Versions/Repositories. The addon.yml should look like this:

```YAML
---
addonProduct: /NAMESPC/PRODUCTX
addonVersion: 1.2.0
repositories:
  - name: /NAMESPC/COMPONENTA
    branch: v1.2.0
    version: 1.2.0
    commitID: 7d4516e9
  - name: /NAMESPC/COMPONENTB
    branch: v2.0.0
    version: 2.0.0
    commitID: 9f102ffb
```

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
  abapAddonAssemblyKitCheckCVs:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    abapAddonAssemblyKitEndpoint: 'https://myabapAddonAssemblyKitEndpoint.com',
    addonDescriptorFileName: 'addon.yml'
```
