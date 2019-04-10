# commonPipelineEnvironment

## Description

Provides project specific settings.

## Prerequisites

none

## Method details

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
