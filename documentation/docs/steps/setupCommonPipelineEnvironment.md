# setupCommonPipelineEnvironment

## Description

Initializes the [`commonPipelineEnvironment`](commonPipelineEnvironment.md), which is used throughout the complete pipeline.

!!! tip
    This step needs to run at the beginning of a pipeline right after the SCM checkout.
    Then subsequent pipeline steps consume the information from `commonPipelineEnvironment`; it does not need to be passed to pipeline steps explicitly.

## Prerequisites

* A **configuration file** with properties (default location: `.pipeline/config.properties`). The property values are used as default values in many pipeline steps.

## Parameters

| parameter    | mandatory | default                       | possible values |
| ------------ |-----------|-------------------------------|-----------------|
| `script`     | yes       | -                             |                 |
| `configFile` | no        | `.pipeline/config.properties` |                 |

* `script` - The reference to the pipeline script (Jenkinsfile). Normally `this` needs to be provided.
* `configFile` - Property file defining project specific settings.

## Step configuration

none

## Side effects

none

## Exceptions

none

## Example

```groovy
setupCommonPipelineEnvironment script: this
```
