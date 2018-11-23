# transportRequestUploadFile

## Description

Uploads a file to a Transport Request.

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters

| parameter        | mandatory | default                                                | possible values    |
| -----------------|-----------|--------------------------------------------------------|--------------------|
| `script`        | yes       |                                                    |                    |
| `changeDocumentId`        | `SOLMAN` only |                                                    |                    |
| `transportRequestId`| yes   |                                                    |                    |
| `applicationId`  | `SOLMAN` only       |                                                    |                    |
| `filePath`        | yes       |                                                    |                    |
| `changeManagement/credentialsId`  | yes       |                                                    |                    |
| `changeManagement/endpoint`        | yes       |                                                    |                    |
| `changeManagement/git/from`         | no        | `origin/master`                                        |                    |
| `changeManagement/git/to`           | no        | `HEAD`                                                 |                    |
| `changeManagement/changeDocumentLabel`        | no        | `ChangeDocument\s?:`                                   | regex pattern      |
| `changeManagement/transportRequestLabel`        | no        | `TransportRequest\s?:`                                   | regex pattern      |
| `changeManagement/git/format`        | no        | `%b`                                                   | see `git log --help` |
| `changeManagement/type`           | no        | `SOLMAN`                                               | `SOLMAN`, `CTS`    |

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `changeDocumentId` - For type `SOLMAN` only. The id of the change document related to the transport request to release. Typically provided via commit history.
* `transportRequestId` - The id of the transport request to release. Typically provided via commit history.
* `applicationId` - For type `SOLMAN` only. The id of the application.
* `filePath` - The path of the file to upload.
* `changeManagement/credentialsId` - The credentials to connect to the service endpoint (Solution Manager, ABAP System).
* `changeManagement/endpoint` - The service endpoint (Solution Manager, ABAP System).
* `changeManagement/git/from` - The starting point for retrieving the change document id and/or transport request id
* `changeManagement/git/to` - The end point for retrieving the change document id and/or transport request id
* `changeManagement/changeDocumentLabel` - For type `SOLMAN` only. A pattern used for identifying lines holding the change document id.
* `changeManagement/transportRequestLabel` - A pattern used for identifying lines holding the transport request id.
* `changeManagement/type` Where/how the transport request is created (via SAP Solution Manager, ABAP).
* `changeManagement/git/format` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.

## Step configuration

## Return value

None.

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

The step is configured using a customer configuration file provided as
resource in an custom shared library.


```groovy
@Library('piper-library-os@master') _

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

The parameters can also be provided when the step is invoked.
