# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (Technical Communication User) must be stored in the Jenkins Credential Store
* This step needs the names of the packages which should be registered. For each package a SAR archive with the data file and metadata XML must be provided.
* The package names and their status are taken from the addonDescriptor in the commonPipelineEnvironment, as well as the SarXMLFilePath with the path to the SAR file.
* The information will be written to the commonPipelineEnvironment if you run prior to this step the step [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages)
* The SAR archive is produced if you run the step [abapEnvironmentAssemblePackages](https://sap.github.io/jenkins-library/steps/abapEnvironmentAssemblePackages)

A detailed description of all prerequisites of the scenario and how to configure them can be found in the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitRegisterPackages script: this
```

If the step is to be configured individually the config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitRegisterPackages:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId'
```

More convenient ways of configuration (e.g. on stage level) are described in the respective scenario/pipeline documentation.
