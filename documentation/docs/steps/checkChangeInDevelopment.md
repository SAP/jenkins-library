# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* No prerequisites

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
