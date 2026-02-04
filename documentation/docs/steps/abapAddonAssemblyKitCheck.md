# ${docGenStepName}

## ${docGenDescription}

### Artifacts

- addonDescriptorFile (addon.yml)
    The addonDescriptorFile as specified in parameter addonDescriptorFileName is archived as artifact. This is done as this file is the main configuration and usually changed with every run. Thus it simplifies support if the corresponding configuration file is directly accessible in the pipeline.

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
abapAddonAssemblyKitCheck script: this
```

If the step is to be configured individually the config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitCheck:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId',
    addonDescriptorFileName: 'addon.yml'
```

More convenient ways of configuration (e.g. on stage level) are described in the respective scenario/pipeline documentation.
