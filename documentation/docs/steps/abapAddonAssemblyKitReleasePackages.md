# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (Technical Communication User) must be stored in the Jenkins Credential Store
* This step needs the names of the packages which should be released. The packages needs to be in status "L"ocked. If they are already in status "R"eleased it is fine, then the release will just not be executed. However this step will end with an error if a package has status "P"lanned.
* The package names are taken from the addonDescriptor in the commonPipelineEnvironment together with the status of the packages.
* The step [abapAddonAssemblyKitRegisterPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitRegisterPackages) will set the status of the packages to "L"ocked and writes the needed data to the commonPipelineEnvironment.

A detailed description of all prerequisites of the scenario and how to configure them can be found in the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitReleasePackages script: this
```

If the step is to be configured individually the config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitReleasePackages:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId'
```

More convenient ways of configuration (e.g. on stage level) are described in the respective scenario/pipeline documentation.
