# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

You need to have a Bats test file. By default you would put this into directory `src/test` within your source code repository.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
batsExecuteTests script:this
testsPublishResults junit: [pattern: '**/Test-*.xml', archive: true]
```

    With `envVars` it is possible to pass either fixed values but also templates using [`commonPipelineEnvironment`](commonPipelineEnvironment.md).

    Example:

    ```yaml
    batsExecuteTests script: this, envVars = [
      FIX_VALUE: 'my fixed value',
      CONTAINER_NAME: '\${commonPipelineEnvironment.configuration.steps.executeBatsTests.dockerContainerName}',
      IMAGE_NAME: '\${return commonPipelineEnvironment.getDockerImageNameAndTag()}'
    ]
    ```

    This means within the test one could refer to environment variables by calling e.g.
    `run docker run --rm -i --name \$CONTAINER_NAME --entrypoint /bin/bash \$IMAGE_NAME echo "Test"`
