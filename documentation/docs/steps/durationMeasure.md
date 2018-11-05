# durationMeasure

## Description

This step is used to measure the duration of a set of steps, e.g. a certain stage.
The duration is stored in a Map. The measurement data can then be written to an Influx database using step [influxWriteData](influxWriteData.md).

!!! tip
    Measuring for example the duration of pipeline stages helps to identify potential bottlenecks within the deployment pipeline.
    This then helps to counter identified issues with respective optimization measures, e.g parallelization of tests.

## Prerequisites

none

## Pipeline configuration

none

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | yes | |  |
| measurementName | no | test_duration |  |

Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `measurementName` defines the name of the measurement which is written to the Influx database.

## Step configuration

none

## Example

```groovy
durationMeasure (script: this, measurementName: 'build_duration') {
    //execute your build
}
```
