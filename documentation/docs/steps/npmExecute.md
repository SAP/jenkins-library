# npmExecute

## Description

Executes NPM commands inside a docker container.
Docker image, docker options and npm commands can be specified or configured.

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `defaultNpmRegistry` | no |  |  |
| `dockerImage` | no | `node:8-stretch` |  |
| `dockerOptions` | no |  |  |
| `npmCommand` | no |  |  |
| `script` | yes |  |  |

* `defaultNpmRegistry` - URL of default NPM registry
* `dockerImage` - Name of the docker image that should be used, in which node should be installed and configured. Default value is 'node:8-stretch'.
* `dockerOptions` - Docker options to be set when starting the container.
* `npmCommand` - Which NPM command should be executed.
* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in `script: this`. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections of the config.yml the configuration is possible:

| parameter | general | step | stage |
|-----------|---------|------|-------|
| `defaultNpmRegistry` |  | X | X |
| `dockerImage` |  | X | X |
| `dockerOptions` |  |  | X |
| `npmCommand` |  | X | X |
| `script` |  |  |  |

## Exceptions

None

## Examples

```groovy
npmExecute script: this, dockerImage: 'node:8-stretch', npmCommand: 'run build'
```
