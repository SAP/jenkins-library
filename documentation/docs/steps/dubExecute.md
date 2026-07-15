

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.
# ${docGenStepName}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Exceptions

None

## Examples

```groovy
dubExecute script: this, dockerImage: 'dlang2/dmd-ubuntu:latest', dubCommand: 'build'
```
