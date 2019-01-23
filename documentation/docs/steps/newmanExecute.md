# newmanExecute

## Description

This script executes your [Postman](https://www.getpostman.com) tests from a collection via the [Newman](https://www.getpostman.com/docs/v6/postman/collection_runs/command_line_integration_with_newman) command line collection.

## Prequisites

- prepared Postman with a test collection

## Example

Pipeline step:

```groovy
newmanExecute script: this
```

This step should be used in combination with `testsPublishResults`:

```groovy
newmanExecute script: this, failOnError: false
testsPublishResults script: this, junit: [pattern: '**/newman/TEST-*.xml']
```

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `dockerImage` | no |  |  |
| `failOnError` | no |  | `true`, `false` |
| `gitBranch` | no |  |  |
| `gitSshKeyCredentialsId` | no |  |  |
| `newmanCollection` | no |  |  |
| `newmanEnvironment` | no |  |  |
| `newmanGlobals` | no |  |  |
| `newmanRunCommand` | no |  |  |
| `script` | yes |  |  |
| `stashContent` | no |  |  |
| `testRepository` | no |  |  |

- `dockerImage` - Docker image for code execution.
- `failOnError` - Defines the behavior, in case tests fail.
- `gitBranch` - see `testRepository`
- `gitSshKeyCredentialsId` - see `testRepository`
- `newmanCollection` - The test collection that should be executed. This could also be a file pattern.
- `newmanEnvironment` - Specify an environment file path or URL. Environments provide a set of variables that one can use within collections. see also [Newman docs](https://github.com/postmanlabs/newman#newman-run-collection-file-source-options)
- `newmanGlobals` - Specify the file path or URL for global variables. Global variables are similar to environment variables but have a lower precedence and can be overridden by environment variables having the same name. see also [Newman docs](https://github.com/postmanlabs/newman#newman-run-collection-file-source-options)
- `newmanRunCommand` - The newman command that will be executed inside the docker container.
- `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in script: this. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.
- `stashContent` - If specific stashes should be considered for the tests, you can pass this via this parameter.
- `testRepository` - In case the test implementation is stored in a different repository than the code itself, you can define the repository containing the tests using parameter `testRepository` and if required `gitBranch` (for a different branch than master) and `gitSshKeyCredentialsId` (for protected repositories). For protected repositories the `testRepository` needs to contain the ssh git url.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:| parameter | general | step | stage |
|-----------|---------|------|-------|
| `dockerImage` |  | X | X |
| `failOnError` |  | X | X |
| `gitBranch` |  | X | X |
| `gitSshKeyCredentialsId` |  | X | X |
| `newmanCollection` |  | X | X |
| `newmanEnvironment` |  | X | X |
| `newmanGlobals` |  | X | X |
| `newmanRunCommand` |  | X | X |
| `script` | X | X | X |
| `stashContent` |  | X | X |
| `testRepository` |  | X | X |
