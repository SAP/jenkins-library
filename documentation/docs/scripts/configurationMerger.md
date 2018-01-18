# ConfigurationMerger

## Description
A helper script that can merge the configurations from multiple sources.

## Static Method Details

### merge

#### Description

A step is usually configured by default values, configuration values from the configuration file and the parameters.
The method can merge these sources.
Default values are overwritten by configuration file values.
These are overwritten by parameters.

#### Parameters

| parameter          | mandatory | Class                             |
| -------------------|-----------|-----------------------------------|
| `parameters`       | yes       | Map                               |
| `parameterKeys`    | yes       | List                              |
| `configurationMap` | yes       | Map                               |
| `configurationKeys`| yes       | List                              |
| `defaults`         | yes       | Map                               |

* `parameters` Parameters map given to the step
* `parameterKeys` List of parameter names (keys) that should be considered while merging.
* `configurationMap` Configuration map loaded from the configuration file.
* `configurationKeys` List of configuration keys that should be considered while merging.
* `defaults` Map of default values, e.g. loaded from the default value configuration file.

#### Side effects

none

#### Example

```groovy
prepareDefaultValues script: script
final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, 'mavenExecute')

final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, 'mavenExecute')

List parameterKeys = [
    'dockerImage',
    'globalSettingsFile',
    'projectSettingsFile',
    'pomPath',
    'flags',
    'goals',
    'm2Path',
    'defines'
]
List stepConfigurationKeys = [
    'dockerImage',
    'globalSettingsFile',
    'projectSettingsFile',
    'pomPath',
    'm2Path'
]

Map configuration = ConfigurationMerger.merge(parameters, parameterKeys, stepConfiguration, stepConfigurationKeys, stepDefaults)
```

### mergeWithPipelineData

#### Description

A step is usually configured by default values, configuration values from the configuration file and the parameters.
In certain cases also information previously generated in the pipeline should be mixed in, like for example an artifactVersion created earlier.
The method can merge these sources.
Default values are overwritten by configuration file values.
Those are overwritten by information previously generated in the pipeline (e.g. stored in [commonPipelineEnvironment](../steps/commonPipelineEnvironment.md)).
These are overwritten by parameters passed directly to the step.

#### Parameters

| parameter          | mandatory | Class                             |
| -------------------|-----------|-----------------------------------|
| `parameters`       | yes       | Map                               |
| `parameterKeys`    | yes       | List                              |
| `pipelineDataMap`  | yes       | Map                               |
| `configurationMap` | yes       | Map                               |
| `configurationKeys`| yes       | List                              |
| `defaults`         | yes       | Map                               |

* `parameters` Parameters map given to the step
* `parameterKeys` List of parameter names (keys) that should be considered while merging.
* `configurationMap` Configuration map loaded from the configuration file.
* `pipelineDataMap` Values available to the step during pipeline run.
* `configurationKeys` List of configuration keys that should be considered while merging.
* `defaults` Map of default values, e.g. loaded from the default value configuration file.

#### Side effects

none

#### Example

```groovy
def stepName = 'influxWriteData'
prepareDefaultValues script: script

final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, stepName)
final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)
final Map generalConfiguration = ConfigurationLoader.generalConfiguration(script)

List parameterKeys = [
    'artifactVersion',
    'influxServer',
    'influxPrefix'
]
Map pipelineDataMap = [
    artifactVersion: commonPipelineEnvironment.getArtifactVersion()
]
List stepConfigurationKeys = [
    'influxServer',
    'influxPrefix'
]

Map configuration = ConfigurationMerger.mergeWithPipelineData(parameters, parameterKeys, pipelineDataMap, stepConfiguration, stepConfigurationKeys, stepDefaults)
```
