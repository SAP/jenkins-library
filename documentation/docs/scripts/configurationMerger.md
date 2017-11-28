# ConfigurationMerger

## Description
A helper script that can merge the configurations from multiple sources. 
A step is usually configured by default values, configuration values from the configuration file and the parameters.
The helper can merge these maps. 
Default values are overwritten by configuration file values. 
These are overwritten by parameters. 

## Static Method Details

### merge

#### Description

Returns the configuration for a specific step as map.

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
* `configurationKeys` List of configuration (keys) that should be considered while merging. 
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
