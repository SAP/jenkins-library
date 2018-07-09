# checkChangeInDevelopment

## Description
Checks if a Change Document is in status 'in development'. The change document id is retrieved from the git commit history. The change document id
can also be provided via parameter `changeDocumentId`. Any value provided as parameter has a higher precedence than a value from the commit history.

By default the git commit messages between `origin/master` and `HEAD` are scanned for a line like `ChangeDocument : <changeDocumentId>`. The commit
range and the pattern can be configured. For details see 'parameters' table. 

## Prerequisites
* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters
| parameter          | mandatory | default                                                | possible values    |
| -------------------|-----------|--------------------------------------------------------|--------------------|
| `script`           | yes       |                                                        |                    |
| `changeDocumentId` | yes       |                                                        |                    |
| `credentialsId`    | yes       |                                                        |                    |
| `endpoint`         | yes       |                                                        |                    |
| `gitFrom`         | no        | `origin/master`                                        |                    |
| `gitTo`           | no        | `HEAD`                                                 |                    |
| `gitChangeDocumentLabel`        | no        | `ChangeDocument\s?:`                                   | regex pattern      |

* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving, for example, configuration parameters.
* `changeDocumentId` - The id of the change document to transport. If not provided, it is retrieved from the git commit history.
* `credentialsId` - The credentials to connect to the Solution Manager.
* `endpoint` - The address of the Solution Manager.
* `gitFrom` - The starting point for retrieving the change document id
* `gitTo` - The end point for retrieving the change document id
* `gitChangeDocumentLabel` - A pattern used for identifying lines holding the change document id.

## Step configuration
The following parameters can also be specified as step parameters using the global configuration file:

* `credentialsId`
* `endpoint`

## Return value
`true` in case the change document is in status 'in development'. Otherwise an hudson.AbortException is thrown. In case `failIfStatusIsNotInDevelopment`
is set to `false`, `false` is returned in case the change document is not in status 'in development'

## Exceptions
* `AbortException`:
    * If the change id is not provided via parameter and if the change document id cannot be retrieved from the commit history.
    * If the change is not in status `in development`. In this case no exception will be thrown when `failIfStatusIsNotInDevelopment` is set to `false`.

## Example
```groovy
    checkChangeInDevelopment script:this
```

