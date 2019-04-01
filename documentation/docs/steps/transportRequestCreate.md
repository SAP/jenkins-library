# transportRequestCreate

## Description

Creates

* a Transport Request for a Change Document on the Solution Manager (type `SOLMAN`) or
* a Transport Request inside an ABAP system (type`CTS`)

The id of the transport request is availabe via [commonPipelineEnvironment.getProperty('transportRequestId')](commonPipelineEnvironment.md)

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.
* Solution Manager version `ST720 SP08` or newer.

## Parameters

| parameter        | mandatory | default                                                | possible values    |
| -----------------|-----------|--------------------------------------------------------|--------------------|
| `script`        | yes       |                                                    |                    |
| `changeDocumentId`        | for `SOLMAN`      |                                                    |                    |
| `transportType`  | for `CTS`  | no                                                    |                    |
| `targetSystem`   | for `CTS`  | no                                                    |                    |
| `description`    | for `CTS`  | no                                                    |                    |
| `changeManagement/credentialsId`  | yes       |                                                    |                    |
| `changeManagement/endpoint`        | yes       |                                                    |                    |
| `changeManagement/clientOpts`     | no       |                                                     |                     |
| `changeManagement/git/from`         | no        | `origin/master`                                        |                    |
| `changeManagement/git/to`           | no        | `HEAD`                                                 |                    |
| `changeManagement/changeDocumentLabel`        | no        | `ChangeDocument\s?:`                                   | regex pattern      |
| `changeManagement/git/format`        | no        | `%b`                                                   | see `git log --help` |
| `changeManagement/type`           | no        | `SOLMAN`                                               | `SOLMAN`, `CTS`    |
| `developmentSystemId` | for `SOLMAN` |         | |

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `changeDocumentId` - for `SOLMAN` only. The id of the change document to that the transport request is bound to. Typically this value is provided via commit message in the commit history.
* `changeManagement/type` Where/how the transport request is created (via SAP Solution Manager, ABAP).
* `changeManagement/credentialsId` - The credentials to connect to the service endpoint (Solution Manager, ABAP System).
* `changeManagement/endpoint` - The service endpoint (Solution Manager, ABAP System).
* `changeManagement/clientOpts`- Options forwarded to JVM used by the CM client, like `JAVA_OPTS`
* `changeManagement/git/from` - The starting point for retrieving the change document id
* `changeManagement/git/to` - The end point for retrieving the change document id
* `changeManagement/changeDocumentLabel` - For type `SOLMAN` only. A pattern used for identifying lines holding the change document id.
* `changeManagement/git/format` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
* `description` - for `CTS` only. The description of the transport request.
* `targetSystem` - for `CTS` only. The system receiving the transport request.
* `transportType` - for type `CTS` only. Typically `W` (workbench) or `C` customizing.
* `developmentSystemId`- for `SOLMAN` only. The logical system id for which the transport request is created. The format is `<SID>~<TYPE>(/<CLIENT>)?`. For ABAP Systems the `developmentSystemId` looks like `DEV~ABAP/100`. For non-ABAP systems the `developmentSystemId` looks like e.g. `L21~EXT_SRV` or `J01~JAVA`. In case the system type is not known (in the examples provided here: `EXT_SRV` or `JAVA`) the information can be retrieved from the Solution Manager instance.
## Step configuration

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
