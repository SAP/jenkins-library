# transportRequestCreate

## Description

Creates

* a Transport Request for a Change Document on the Solution Manager (type `SOLMAN`) or
* a Transport Request inside an ABAP system (type`CTS`)

The id of the transport request is availabe via [commonPipelineEnvironment.getTransportRequestId()](commonPipelineEnvironment.md)

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `changeDocumentId` | `SOLMAN` only, can be provided via git commit history. |  |  |
| `changeManagement/changeDocumentLabel` | `SOLMAN` only | `ChangeDocument\s?:` | regex pattern |
| `changeManagement/clientOpts` | yes |  |  |
| `changeManagement/credentialsId` | yes | `CM` |  |
| `changeManagement/endpoint` | yes |  |  |
| `changeManagement/git/format` | yes | `%b` | see `git log --help` |
| `changeManagement/git/from` | yes | `origin/master` |  |
| `changeManagement/git/to` | yes | `HEAD` |  |
| `changeManagement/type` | yes | `NONE` | `SOLMAN`, `CTS`, `NONE` |
| `description` | `CTS` only |  |  |
| `developmentSystemId` | `SOLMAN` only. |  |  |
| `script` | yes |  |  |
| `targetSystem` | `CTS` only |  |  |
| `transportType` | `CTS` only |  |  |

* `changeDocumentId` - for `SOLMAN` only. The id of the change document to that the transport request is bound to. Typically this value is provided via commit message in the commit history.
* `changeManagement/changeDocumentLabel` - For type `SOLMAN` only. A pattern used for identifying lines holding the change document id.
* `changeManagement/clientOpts` - Options forwarded to JVM used by the CM client, like `JAVA_OPTS`.
* `changeManagement/credentialsId` - The credentials to connect to the service endpoint (Solution Manager, ABAP System).
* `changeManagement/endpoint` - The service endpoint (Solution Manager, ABAP System).
* `changeManagement/git/format` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
* `changeManagement/git/from` - The starting point for retrieving the change document id.
* `changeManagement/git/to` - The end point for retrieving the change document id.
* `changeManagement/type` - Where/how the transport request is created (via SAP Solution Manager, ABAP).
* `description` - For type `CTS` only. The description of the transport request.
* `developmentSystemId` - for `SOLMAN` only. Outlines how the artifact is handled. For CTS use case: `SID~Type/Client`, e.g. `XX1~ABAP/100`, for SOLMAN use case: `SID~Typ`, e.g. `J01~JAVA`.
* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in `script: this`. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.
* `targetSystem` - For type `CTS` only. The system receiving the transport request.
* `transportType` - for type `CTS` only. Typically `W` (workbench) or `C` customizing.


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
| `changeManagement/type` |  | X | X |
| `description` |  | X | X |
| `developmentSystemId` |  | X | X |
| `script` | X | X | X |
| `targetSystem` |  | X | X |
| `transportType` |  | X | X |


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

The properties configured in section `'general/changeManagement'` are shared between
all change managment related steps.

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

The parameters can also be provided when the step is invoked.
