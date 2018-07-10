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

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `changeDocumentId` - The id of the change document to transport.
* `credentialsId` - The credentials to connect to the Solution Manager.
* `endpoint` - The address of the Solution Manager.
* `clientOpts`- Options forwarded to JVM used by the CM client, like `JAVA_OPTS`

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

