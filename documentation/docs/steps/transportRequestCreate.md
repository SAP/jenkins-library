# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* For type SOLMAN

| SOLMAN Version | cm client version |
|-----------------------------------------|----------------|
|SAP Solution Manager 7.2 SP6, SP7        | cm_client v1.x |
|SAP Solution Manager 7.2 SP 8 and higher | cm_client v2.0 |

* For type CTS (without SOLMAN)

| AS ABAP Version       |     Service Pack    |
|-----------------------|---------------------|
| 7.50                  |  >= SP12            |
| 7.51                  |  >= SP07            |
| 7.52                  |  >= SP03            |
| 7.53                  |  >= SP01            |
| 7.54                  |  >= SP01            |

CM Client version v2.x is preconfigured. In order to switch to
version v1.x update `changeManagement/cts/docker/image` with
`ppiper/cm-client:1.0.0.0`.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

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
    type: 'SOLMAN'
    endpoint: 'https://example.org/cm'
    git:
      from: 'HEAD~1'
      to: 'HEAD'
      format: '%b'
```

The properties configured in section `'general/changeManagement'` are shared between
all change management related steps.

The properties can also be configured on a per-step basis:

```yaml
  [...]
  steps:
    transportRequestCreate:
      changeManagement:
        type: 'SOLMAN'
        endpoint: 'https://example.org/cm'
        [...]
```

The parameters can also be provided when the step is invoked. For examples see below.

## Return value

none

## Exceptions

* `AbortException`:
  * If the creation of the transport request fails.
* `IllegalStateException`:
  * If the change id is not provided.

## Example

```groovy
// SOLMAN
def transportRequestId = transportRequestCreate script:this,
                                                changeDocumentId: '001,'
                                                changeManagement: [
                                                  type: 'SOLMAN'
                                                  endpoint: 'https://example.org/cm'
                                                ]
// CTS
def transportRequestId = transportRequestCreate script:this,
                                                transportType: 'W',
                                                targetSystem: 'XYZ',
                                                description: 'the description',
                                                changeManagement: [
                                                  type: 'CTS'
                                                  endpoint: 'https://example.org/cm'
                                                ]
```
