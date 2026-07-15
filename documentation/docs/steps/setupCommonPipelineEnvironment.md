# ${docGenStepName}

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.

## ${docGenDescription}

## Prerequisites

* A **configuration file** with properties. The property values are used as default values in many pipeline steps.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Exceptions

none

## Example

```groovy
setupCommonPipelineEnvironment script: this
```
