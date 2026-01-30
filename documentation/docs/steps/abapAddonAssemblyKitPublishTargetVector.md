# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (Technical Communication User) must be stored in the Jenkins Credential Store
* This step needs an existing Target Vector as well as the scope where it should be published.
* The Target Vector ID is taken from the addonDescriptor in the commonPipelineEnvironment.
* If you run prior to this step the step [abapAddonAssemblyKitCreateTargetVector](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCreateTargetVector), the Target Vector will be created and its ID will be written to the commonPipelineEnvironment

A detailed description of all prerequisites of the scenario and how to configure them can be found in the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitPublishTargetVector(
                    targetVectorScope: 'T',
                    script: this,
                    )
```

If the step is to be configured individually the config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitPublishTargetVector:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId'
```

More convenient ways of configuration (e.g. on stage level) are described in the respective scenario/pipeline documentation.
