# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* For type SOLMAN

| SOLMAN Version | cm client version |
|-----------------------------------------|----------------|
|SAP Solution Manager 7.2 SP6, SP7        | cm_client v1.x |
|SAP Solution Manager 7.2 SP 8 and higher | cm_client v2.0 |

CM Client version v2.x is preconfigured. In order to switch to
version v1.x update `changeManagement/cts/docker/image` with
`ppiper/cm-client:1.0.0.0`.

* For type CTS (without SOLMAN)

| AS ABAP Version       |     Service Pack    |
|-----------------------|---------------------|
| 7.50                  |  >= SP12            |
| 7.51                  |  >= SP07            |
| 7.52                  |  >= SP03            |
| 7.53                  |  >= SP01            |
| 7.54                  |  >= SP01            |

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
