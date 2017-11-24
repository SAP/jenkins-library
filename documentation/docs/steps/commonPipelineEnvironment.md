# commonPipelineEnvironment

## Description
Provides project specific settings.

## Prerequisites
none


## Method details

### getConfigProperties()

#### Description
Returns the map of project specific configuration properties. No defensive copy is created.
Write operations to the map are visible further down in the pipeline.

#### Parameters
none

#### Return value
A map containing project specific configuration properties.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.getConfigProperties()
```


### setConfigProperties(configuration)

#### Description
Sets the map of configuration properties. Any existing map is overwritten.

#### Parameters
* `configuration` - A map containing the new configuration

#### Return value
none

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setConfigProperties([DEPLOY_HOST: 'deploy-host.com', DEPLOY_ACCOUNT: 'deploy-account'])
```


### getConfigProperty(property)

#### Description
Gets a specific value from the configuration property.

#### Parameters
* `property` - The key of the property.

#### Return value
* The value associated with key `property`. `null` is returned in case the property does not exist.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.getConfigProperty('DEPLOY_HOST')
```


### setConfigProperty(property, value)

#### Description
Sets property `property` with value `value`. Any existing property with key `property` is overwritten.

#### Parameters
* `property` - The key of the property.
* `value` - The value of the property.

#### Return value
none

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'my-deploy-host.com')
```


### getMtarFileName()

#### Description
Returns the path of the mtar archive file.

#### Parameters
none

#### Return value
The path of the mtar archive file.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.getMtarFileName()
```

### setMtarFileName(name)

#### Description
Sets the path of the mtar archive file. Any old value is discarded.

#### Parameters
* `mtarFilePath` - The path of the mtar archive file name.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setMtarFileName('path/to/foo.mtar')
```
