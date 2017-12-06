# ConfigurationLoader

## Description
Loads configuration values from the global configuration. 
The global configuration is stored in the commonPipelineEnvironment and should be loaded before by calling setupCommonPipelineEnvironment.

## Static Method Details

### stepConfiguration

#### Description

Returns the configuration for a specific step as map.

#### Parameters

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `stepName` - The name of the step

#### Side effects

none

#### Example

In your `config.yml` you define the following:

```
#Steps Specific Configuration
steps:
  mavenExecute:
    dockerImage: 'maven:3.5-jdk-7'
```

To get the map containing the key `dockerImage` and the value `maven:3.5-jdk-7` you have to execute the following:

```groovy
Map configuration = ConfigurationLoader.stepConfiguration(script, 'mavenExecute')
```

### defaultStepConfiguration

#### Description

Returns the default configuration for a specific step as map.

#### Parameters

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `stepName` - The name of the step

#### Side effects

none

#### Example

To get the map of the default values defined in the file `resources/default_pipeline_environment.yml` you have to execute the following:

```groovy
Map configuration = ConfigurationLoader.defaultStepConfiguration(script, 'mavenExecute')
```
### generalConfiguration

#### Description

Returns the configuration in the section general of the configuration file.

#### Parameters

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.

#### Side effects

none

#### Example

In your `config.yml` you define the following:

```
#Project Setup
general:
  productiveBranch: 'master'
```

To get the map containing the key `productiveBranch` and the value `master` you have to execute the following:

```groovy
Map configuration = ConfigurationLoader.generalConfiguration(script)
```

### defaultGeneralConfiguration

#### Description

Returns the default configuration in the section general of the default configuration file.

#### Parameters

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.

#### Side effects

none

#### Example

To get the map of the default values defined in the file `resources/default_pipeline_environment.yml` you have to execute the following:

```groovy
Map configuration = ConfigurationLoader.defaultGeneralConfiguration(script)
```

### stageConfiguration

#### Description

Returns the configuration for a specific stage as map.
This is useful if you decide to have a central pipeline and want to give all your projects the possibility to configure the stages in the central pipeline.
Thus, the central pipeline can define how to deploy and read the configuration.
In the their configuration files, all the projects can configure the location where to deploy. 

#### Parameters

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `script` - Name of the stage as defined in the configuration file.

#### Side effects

none

#### Example

In your `config.yml` you define the following:

```
#Project Setup
#Stage Specific Configurations
stages:
  productionDeployment:
    targets:
      - apiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
        org: 'myOrg'
        manifest: 'manifest.yml'
        appName: 'my-app'
```

To get the map containing the key `targets` and the list of the deployment locations.

```groovy
Map configuration = ConfigurationLoader.stageConfiguration(script, 'productionDeployment')
```
