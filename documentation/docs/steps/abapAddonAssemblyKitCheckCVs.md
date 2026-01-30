# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (Technical Communication User) must be stored in the Jenkins Credential Store
* The step needs an addon.yml containing information about the Product Version and corresponding Software Component Versions/Repositories

A detailed description of all prerequisites of the scenario and how to configure them can be found in the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitCheckPV script: this
```

If the step is to be configured individually the config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitCheckCVs:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    addonDescriptorFileName: 'addon.yml'
```

More convenient ways of configuration (e.g. on stage level) are described in the respective scenario/pipeline documentation.
