# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## ${docGenParameters}

## ${docGenConfiguration}

The step is configured using a customer configuration file provided as
resource in an custom shared library.

```groovy
@Library('piper-lib-os@master') _

// the shared lib containing the additional configuration
// needs to be configured in Jenkins
@Library('foo@master') __

// inside the shared lib denoted by 'foo' the additional configuration file
// needs to be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this,
                             customDefaults: 'myConfig.yml'
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

The properties configured in section `'general/changeManagement'` are shared between all change managment related steps.

The properties can also be configured on a per-step basis:

```yaml
  [...]
  steps:
    transportRequestUploadFile:
      applicationId: 'FOO'
      changeManagement:
        type: 'SOLMAN'
        endpoint: 'https://example.org/cm'
        [...]
```

The parameters can also be provided when the step is invoked. For examples see below.

## Exceptions

* `IllegalArgumentException`:
  * If the change id is not provided (`SOLMAN` only).
  * If the transport request id is not provided.
  * If the application id is not provided (`SOLMAN` only).
  * If the file path is not provided.
* `AbortException`:
  * If the upload fails.

## Example

```groovy
// SOLMAN
transportRequestUploadFile script:this,
                           changeDocumentId: '001',   // typically provided via git commit history
                           transportRequestId: '001', // typically provided via git commit history
                           applicationId: '001',
                           filePath: '/path',
                           changeManagement:[
                             type: 'SOLMAN'
                             endpoint: 'https://example.org/cm'
                           ]
// CTS
transportRequestUploadFile script:this,
                           transportRequestId: '001', // typically provided via git commit history
                           filePath: '/path',
                           changeManagement:[
                             type: 'CTS'
                             endpoint: 'https://example.org/cm'
                           ]
```
