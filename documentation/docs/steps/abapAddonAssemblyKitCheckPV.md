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