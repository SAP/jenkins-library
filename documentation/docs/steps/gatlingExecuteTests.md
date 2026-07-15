# ${docGenStepName}

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.

## ${docGenDescription}

## Prerequisites

The [Gatling Jenkins plugin](https://plugins.jenkins.io/gatling/) needs to be installed.

## ${docGenParameters}

## ${docGenConfiguration}

We recommend to define values of step parameters via [config.yml file](../configuration.md).

## ${docJenkinsPluginDependencies}

## Example

Pipeline step:

```groovy
gatlingExecuteTests script: this, pomPath: 'performance-tests/pom.xml'
```
