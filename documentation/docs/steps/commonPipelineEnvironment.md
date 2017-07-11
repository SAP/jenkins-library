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
Sets the map of configuration properties. An existing map is overwritten.

#### Parameters
* configuration - A map containing the new configuration

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

### getConfigProperty(key)

#### Description
Gets a specific value from the configuration property.

#### Parameters
* key - The key of the property.

#### Return value
* The value associated with key `key`. `null` is returned in case the property does not exist.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.getConfigProperty('DEPLOY_HOST')
```

### setConfigProperty(key, value)

#### Description
Sets property `key` with value `value`. Any existing property with key `key`is overwritten.

#### Parameters
* `key` The key
* `value` The value

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
Returns the name of the mtar file.

#### Parameters
none

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
Sets the name of the mtar file. Any old value is discarded.

#### Parameters
The name of the mtar file name.

#### Side effects
none

#### Exceptions
none

#### Example

```groovy
commonPipelineEnvironment.setMtarFileName('foo')
```
