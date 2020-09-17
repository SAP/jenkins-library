# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (e.g. S-User) must be stored in the Jenkins Credential Store
* The step needs an addon.yml containing information about the Product Version and corresponding Software Component Versions/Repositories. The addon.yml should look like this:

```yaml
addonProduct: /DMO/myAddonProduct
addonVersion: 3.1.4
addonUniqueID: myAddonId
customerID: $ID
repositories:
    - name: /DMO/REPO_A
      tag: v-1.0.1-build-0001
      version: 1.0.1
    - name: /DMO/REPO_B
      tag: rel-2.1.1-build-0001
      version: 2.1.1
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
  abapAddonAssemblyKitCheckPV:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    abapAddonAssemblyKitEndpoint: 'https://myabapAddonAssemblyKitEndpoint.com',
    addonDescriptorFileName: 'addon.yml'
```
