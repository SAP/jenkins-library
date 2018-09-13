# transportRequestRelease

## Description
Releases a Transport Request for a Change Document on the Solution Manager.

## Prerequisites
* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters
| parameter        | mandatory | default                                                | possible values    |
| -----------------|-----------|--------------------------------------------------------|--------------------|
| `script`        | yes       |                                                    |                    |
| `changeDocumentId`        | yes       |                                                    |                    |
| `transportRequestId`| yes   |                                                    |                    |
| `changeManagement/changeDocumentLabel`        | no        | `ChangeDocument\s?:`                                   | regex pattern      |
| `changeManagment/transportRequestLabel`        | no        | `TransportRequest\s?:`                                   | regex pattern |
| `changeManagement/credentialsId`    | yes       |                                                        |                    |
| `changeManagement/endpoint`         | yes       |                                                        |                    |
| `changeManagement/git/from`         | no        | `origin/master`                                        |                    |
| `changeManagement/git/to`           | no        | `HEAD`                                                 |                    |
| `changeManagement/git/format`        | no        | `%b`                                                   | see `git log --help` |

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `changeDocumentId` - The id of the change document related to the transport request to release.
* `transportRequestId` - The id of the transport request to release.
* `changeManagement/changeDocumentLabel` - A pattern used for identifying lines holding the change document id.
* `changeManagment/transportRequestLabel` - A pattern used for identifying lines holding the transport request id.
* `changeManagement/credentialsId` - The id of the credentials to connect to the Solution Manager. The credentials needs to be maintained on Jenkins.
* `changeManagement/endpoint` - The address of the Solution Manager.
* `changeManagement/git/from` - The starting point for retrieving the change document id
* `changeManagement/git/to` - The end point for retrieving the change document id
* `changeManagement/git/format` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.

## Step configuration
The step is configured using a customer configuration file provided as
resource in an custom shared library.

```
@Library('piper-library-os@master') _

// the shared lib containing the additional configuration
// needs to be configured in Jenkins
@Library(foo@master') __

// inside the shared lib denoted by 'foo' the additional configuration file
// needs to be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this,
                             customDefaults: 'myConfig.yml'
```

Example content of ```'resources/myConfig.yml'``` in branch ```'master'``` of the repository denoted by
```'foo'```:

```
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

The properties configured in section `'general/changeManagement'` are shared between all change managment related steps.

The properties can also be configured on a per-step basis:

```
  [...]
  steps:
    transportRequestRelease:
      changeManagement:
        endpoint: 'https://example.org/cm'
        [...]
```

The parameters can also be provided when the step is invoked. For examples see below.

## Exceptions
* `IllegalArgumentException`:
    * If the change id is not provided.
    * If the transport request id is not provided.
* `AbortException`:
    * If the release of the transport request fails.

## Example
```groovy
transportRequestRelease script:this,
                        changeDocumentId: '001',
                        transportRequestId: '001',
                        changeManagement: [
                          endpoint: 'https://example.org/cm'
                        ]
```

