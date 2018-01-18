# commonPipelineEnvironment

## Description
Provides project specific settings.

## Prerequisites
none

## Method details

### getArtifactVersion()

#### Description
Returns the version of the artifact which is build in the pipeline.

#### Parameters
none

#### Return value
A `String` containing the version.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
def myVersion = commonPipelineEnvironment.getArtifactVersion()
```

### setArtifactVersion(version)

#### Description
Sets the version of the artifact which is build in the pipeline.

#### Parameters
none

#### Return value
none

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setArtifactVersion('1.2.3')
```

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

### getInfluxCustomData()

#### Description
Returns the Influx custom data which can be collected during pipeline run.

#### Parameters
none

#### Return value
A `Map` containing the data collected.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
def myInfluxData = commonPipelineEnvironment.getInfluxCustomData()
```

### setInfluxCustomData(data)

#### Description
**This is an internal function!**
Stores Influx custom data collected during pipeline run.

#### Parameters
A `Map` containing the data collected.

#### Return value
none

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setInfluxCustomData([datapoint1: 20, datapoint2: 30, datadescription: 'myDescription'])
```

### getInfluxCustomDataMap()

#### Description
Returns the Influx custom data map which can be collected during pipeline run.
It is used for example by step [`influxWriteData`](../steps/influxWriteData.md).
The data map is a map of maps, like `[pipeline_data: [:], my_measurement: [:]]`
Each map inside the map represents a dedicated measurement in the InfluxDB.


#### Parameters
none

#### Return value
A `Map` containing a `Map`s with data collected.

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
def myInfluxDataMap = commonPipelineEnvironment.getInfluxCustomDataMap()
```

### setInfluxCustomDataMap(data)

#### Description
**This is an internal function!**
Stores Influx custom data collected during pipeline run.

#### Parameters
A `Map` containing a `Map`s with data collected.

#### Return value
none

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setInfluxCustomDataMap([measurement1: [datapoint1: 20, datapoint2: 30, datadescription: 'myDescription'], measurement2: [datapoint1:40]])
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

### getPipelineMeasurement(measurementName)

#### Description
Returns the value of a specific pipeline measurement.
The measurements are collected with step [`durationMeasure`](../steps/durationMeasure.md)

#### Parameters
Name of the measurement

#### Return value
Value of the measurement

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
def myMeasurementValue = commonPipelineEnvironment.getPipelineMeasurement('build_stage_duration')
```

### setPipelineMeasurement(measurementName, value)

#### Description
**This is an internal function!**
Sets the value of a specific pipeline measurement.
Please use the step [`durationMeasure`](../steps/durationMeasure.md) in a pipeline, instead.

#### Parameters
Name of the measurement and its value.

#### Return value
none

#### Side effects
none

#### Exceptions
none

#### Example
```groovy
commonPipelineEnvironment.setPipelineMeasurement('build_stage_duration', 2345)
```
