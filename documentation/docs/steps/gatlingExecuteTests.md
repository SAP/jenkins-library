# ${docGenStepName}

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
