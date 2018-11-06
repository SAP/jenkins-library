# batsExecuteTests

## Description

This step executes tests using the [Bash Automated Testing System - bats-core](https://github.com/bats-core/bats-core)

## Prerequsites

You need to have a Bats test file. By default you would put this into directory `src/test` within your source code repository.

## Parameters

| parameter | mandatory | default | possible values |
|-----------|-----------|---------|-----------------|
| script | yes |  |  |
| dockerImage | no | `node:8-stretch` | |
| dockerWorkspace | no |`/home/node`| |
| envVars | no | `[:]` | |
| failOnError | no | `false` | |
| gitBranch | no | | |
| gitSshKeyCredentialsId | no | | |
| outputFormat | no | `junit` | `tap` |
| repository | no | `https://github.com/bats-core/bats-core.git` | |
| stashContent | no | `['tests']` | |
| testPackage | no | `piper-bats` | |
| testPath | no | `src/test`| |
| testRepository | no | | |

Details:

* `outputFormat` defines the format of the test result output. `junit` would be the standard for automated build environments but you could use also the option `tap`.
* For the transformation of the test result to xUnit format the node module **tap-xunit** is used. `dockerImage` and `dockerWorkspace` define the Docker image used for the transformation and `testPackage` defines the name of the test package used in the xUnit result file.
* `testPath` defines either the directory which contains the test files (`*.bats`) or a single file. You can find further details in the [Bats-core documentation](https://github.com/bats-core/bats-core#usage)
* With `failOnError` you can define the behavior, in case tests fail. For example, in case of `outputFormat: 'junit'` you should set it to `false`. Otherwise test results cannot be recorded using the `testsPublishhResults` step afterwards.
* You can pass environment variables to the test execution by defining parameter `envVars`.

    With `envVars` it is possible to pass either fixed values but also templates using [`commonPipelineEnvironment`](commonPipelineEnvironment.md).

    Example:

    ```yaml
    batsExecuteTests script: this, envVars = [
      FIX_VALUE: 'my fixed value',
      CONTAINER_NAME: '${commonPipelineEnvironment.configuration.steps.executeBatsTests.dockerContainerName}',
      IMAGE_NAME: '${return commonPipelineEnvironment.getDockerImageNameAndTag()}'
    ]
    ```

    This means within the test one could refer to environment variables by calling e.g.
    `run docker run --rm -i --name $CONTAINER_NAME --entrypoint /bin/bash $IMAGE_NAME echo "Test"`

* Using parameters `testRepository` the tests can be loaded from another reposirory. In case the tests are not located in the master branch the branch can be specified with `gitBranch`. For protected repositories you can also define the access credentials via `gitSshKeyCredentialsId`. **Note: In case of using a protected repository, `testRepository` should include the ssh link to the repository.**
* The parameter `repository` defines the version of **bats-core** to be used. By default we use the version from the master branch.

## Step configuration

The following parameters can also be specified as step/stage/general parameters using the [global configuration](../configuration.md):

* dockerImage
* dockerWorkspace
* envVars
* failOnError
* gitBranch
* gitSshKeyCredentialsId
* outputFormat
* repository
* stashContent
* testPackage
* testPath
* testRepository

## Example

```groovy
batsExecuteTests script:this
testsPublishResults junit: [pattern: '**/Test-*.xml', archive: true]
```
