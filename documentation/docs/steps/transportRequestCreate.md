# transportRequestCreate

## Description
Creates a Transport Request for a Change Document on the Solution Manager.

## Prerequisites
* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters
| parameter        | mandatory | default                                                | possible values    |
| -----------------|-----------|--------------------------------------------------------|--------------------|
| `script`        | yes       |                                                    |                    |
| `changeDocumentId`        | yes       |                                                    |                    |
| `credentialsId`  | yes       |                                                    |                    |
| `endpoint`        | yes       |                                                    |                    |
| `clientOpts`     | no       |                                                     |                     |
| `gitFrom`         | no        | `origin/master`                                        |                    |
| `gitTo`           | no        | `HEAD`                                                 |                    |
| `gitChangeDocumentLabel`        | no        | `ChangeDocument\s?:`                                   | regex pattern      |
| `gitFormat`        | no        | `%b`                                                   | see `git log --help` |

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `changeDocumentId` - The id of the change document to transport.
* `credentialsId` - The credentials to connect to the Solution Manager.
* `endpoint` - The address of the Solution Manager.
* `clientOpts`- Options forwarded to JVM used by the CM client, like `JAVA_OPTS`
* `gitFrom` - The starting point for retrieving the change document id
* `gitTo` - The end point for retrieving the change document id
* `gitChangeDocumentLabel` - A pattern used for identifying lines holding the change document id.
* `gitFormat` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.

## Step configuration
The following parameters can also be specified as step parameters using the global configuration file:

* `credentialsId`
* `endpoint`
* `clientOpts`

## Return value
The id of the Transport Request that has been created.

## Exceptions
* `AbortException`:
    * If the change id is not provided.
    * If the creation of the transport request fails.

## Example
```groovy
def transportRequestId = transportRequestCreate script:this, changeDocumentId: '001'
```

