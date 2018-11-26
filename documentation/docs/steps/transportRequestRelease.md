# transportRequestRelease

## Description

Releases a Transport Request.

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `changeDocumentId` | `SOLMAN` only |  |  |
| `changeManagement/changeDocumentLabel` | yes | `ChangeDocument\s?:` | regex pattern |
| `changeManagement/clientOpts` | yes |  |  |
| `changeManagement/credentialsId` | yes | `CM` |  |
| `changeManagement/endpoint` | yes |  |  |
| `changeManagement/git/format` | yes | `%b` |  |
| `changeManagement/git/from` | yes | `origin/master` |  |
| `changeManagement/git/to` | yes | `HEAD` |  |
| `changeManagement/transportRequestLabel` | yes | `TransportRequest\s?:` | regex pattern |
| `changeManagement/type` | yes | `NONE` | `SOLMAN`, `CTS`, `NONE` |
| `script` | yes |  |  |
| `transportRequestId` | yes |  |  |

* `changeDocumentId` - for `SOLMAN` only. The id of the change document related to the transport request to release.
* `changeManagement/changeDocumentLabel` - For type `SOLMAN` only. A pattern used for identifying lines holding the change document id.
* `changeManagement/clientOpts` - Options forwarded to JVM used by the CM client, like `JAVA_OPTS`.
* `changeManagement/credentialsId` - The credentials to connect to the service endpoint (Solution Manager, ABAP System).
* `changeManagement/endpoint` - The service endpoint (Solution Manager, ABAP System).
* `changeManagement/git/format` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
* `changeManagement/git/from` - The starting point for retrieving the change document id and/or transport request id
* `changeManagement/git/to` - The end point for retrieving the change document id and/or transport request id.
* `changeManagement/transportRequestLabel` - A pattern used for identifying lines holding the transport request id.
* `changeManagement/type` - Where/how the transport request is created (via SAP Solution Manager, ABAP).
* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in `script: this`. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.
* `transportRequestId` - The id of the transport request to release.


## Step configuration


We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
|-----------|---------|------|-------|
| `changeDocumentId` |  |  | X |
| `changeManagement/changeDocumentLabel` |  | X | X |
| `changeManagement/clientOpts` |  | X | X |
| `changeManagement/credentialsId` |  | X | X |
| `changeManagement/endpoint` |  | X | X |
| `changeManagement/git/format` |  | X | X |
| `changeManagement/git/from` |  | X | X |
| `changeManagement/git/to` |  | X | X |
| `changeManagement/transportRequestLabel` |  | X | X |
| `changeManagement/type` |  | X | X |
| `script` | X | X | X |
| `transportRequestId` |  |  | X |


## Return value

None.

## Exceptions

* `IllegalArgumentException`:
  * If the change id is not provided (`SOLMAN` only)
  * If the transport request id is not provided.
* `AbortException`:
  * If the release of the transport request fails.

## Example

```groovy
// SOLMAN
transportRequestRelease script:this,
                        changeDocumentId: '001',
                        transportRequestId: '001',
                        changeManagement: [
                          type: 'SOLMAN'
                          endpoint: 'https://example.org/cm'
                        ]
// CTS
transportRequestRelease script:this,
                        transportRequestId: '001',
                        changeManagement: [
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
    transportRequestRelease:
      changeManagement:
        type: 'SOLMAN'
        endpoint: 'https://example.org/cm'
        [...]
```

The parameters can also be provided when the step is invoked.

