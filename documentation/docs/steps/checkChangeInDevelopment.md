# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Migration Guide

**Note:** This step has been deprecated. Use the new step [isChangeInDevelopment](isChangeInDevelopment.md) instead.

Adjust your parameters to the naming convention of the new step.
Adjust the unsupported parameters as indicated in the table below:

| Unsupported Parameter | New Parameter |
| ------------- | ------------- |
| changeManagement/type | This parameter has been removed. `SOLMAN` is the only backend type supported. |
| changeManagement/`<type>`/docker/envVars | `dockerEnvVars` |
| changeManagement/`<type>`/docker/image | `dockerImage` |
| changeManagement/`<type>`/docker/options | `dockerOptions` |
| changeManagement/`<type>`/docker/pullImage | `dockerPullImage` |
| changeManagement/git/format | This parameter has been removed. Make sure that the IDS of your change document and transport request are part of the Git commit message body. |

Your config.yml file should look as follows:

```yaml
general:

# new naming convention
steps:
  isChangeInDevelopment:
    dockerImage: 'ppiper/cm-client:3.0.0.0'
```

**Note:** The new step does not comprise the retrieval of the change document ID from the Git repository anymore. Use the step [transportRequestDocIDFromGit](transportRequestDocIDFromGit.md) instead.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Exceptions

* `AbortException`:
  * If the change id is not provided via parameter and if the change document id cannot be retrieved from the commit history.
  * If the change is not in status `in development`. In this case no exception will be thrown when `failIfStatusIsNotInDevelopment` is set to `false`.
* `IllegalArgumentException`:
  * If a mandatory property is not provided.

## Examples

The step is configured using a customer configuration file provided as
resource in an custom shared library.

```groovy
@Library('piper-lib-os@master') _

// the shared lib containing the additional configuration
// needs to be configured in Jenkins
@Library('foo@master') __

// inside the shared lib denoted by 'foo' the additional configuration file
// needs to be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this, customDefaults: 'myConfig.yml'
```

Example content of `'resources/myConfig.yml'` in branch `'master'` of the repository denoted by
`'foo'`:

```yaml
general:
  changeManagement:
    changeDocumentLabel: 'ChangeDocument\s?:'
    cmClientOpts: '-Djavax.net.ssl.trustStore=<path to truststore>'
    credentialsId: 'CM'
    endpoint: 'https://example.org/cm'
    git:
      from: 'HEAD~1'
      to: 'HEAD'
      format: '%b'
```

The properties configured in section `'general/changeManagement'` are shared between all change management related steps.

The properties can also be configured on a per-step basis:

```yaml
  [...]
  steps:
    checkChangeInDevelopment:
      changeManagement:
        endpoint: 'https://example.org/cm'
        [...]
      failIfStatusIsNotInDevelopment: true
```

The parameters can also be provided when the step is invoked:

```groovy
    // simple case. All mandatory parameters provided via
    // configuration, changeDocumentId provided via commit
    // history
    checkChangeInDevelopment script:this
```

```groovy
    // explicit endpoint provided, we search for changeDocumentId
    // starting at the previous commit (HEAD~1) rather than on
    // 'origin/master' (the default).
    checkChangeInDevelopment(
      script: this
      changeManagement: [
        endpoint: 'https:example.org/cm'
        git: [
          from: 'HEAD~1'
        ]
      ]
    )
```
