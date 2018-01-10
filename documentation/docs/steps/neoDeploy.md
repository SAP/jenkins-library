# neoDeploy

## Description
Deploys an Application to SAP Cloud Platform (SAP CP) using the SAP Cloud Platform Console Client (Neo Java Web SDK).
    
## Prerequisites
* **SAP CP account** - the account to where the application is deployed.
* **SAP CP user for deployment** - a user with deployment permissions in the given account.
* **Jenkins credentials for deployment** - must be configured in Jenkins credentials with a dedicated Id.

![Jenkins credentials configuration](../images/neo_credentials.png)
    
* **Neo Java Web SDK** - can be downloaded from [Maven Central](http://central.maven.org/maven2/com/sap/cloud/neo-java-web-sdk/). The Neo Java Web SDK
needs to be extracted into the folder provided by `neoHome`. In case this parameters is not provided and there is no NEO_HOME parameter in the environment
`<neoRoot>/tools` needs to be in the `PATH`.

* **Java 8 or higher** - needed by the *Neo-Java-Web-SDK*

## Parameters
| parameter          | mandatory | default                                                                                      | possible values |
| -------------------|-----------|----------------------------------------------------------------------------------------------|-----------------|
| `script`           | yes       |                                                                                              |                 |
| `archivePath`      | yes       |                                                                                              |                 |
| `deployHost`       | no        | `'DEPLOY_HOST'` from `commonPipelineEnvironment`                                             |                 |
| `deployAccount`    | no        | `'CI_DEPLOY_ACCOUNT'` from `commonPipelineEnvironment`                                       |                 |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`                                                                        |                 |
| `neoHome`          | no        | Environment is checked for `NEO_HOME`, <br>otherwise the neo toolset is expected in the path |                 |

* `script` - The common script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `archivePath`- The path to the archive for deployment to SAP CP.
* `deployHost` - The SAP Cloud Platform host to deploy to.
* `deployAccount` - The SAP Cloud Platform account to deploy to.
* `credentialsId` - The Jenkins credentials containing user and password used for SAP CP deployment.
* `neoHome` - The path to the `neo-java-web-sdk` tool used for SAP CP deployment. If no parameter is provided, the path is retrieved from the Jenkins environment variables using `env.NEO_HOME`. If this Jenkins environment variable is not set it is assumed that the tool is available in the `PATH`.

## Return value
none

## Side effects
none

## Exceptions
* `Exception`:
    * If `archivePath` is not provided.
* `AbortException`:
    * If neo-java-web-sdk is not installed, or `neoHome`is wrong.
    * If `deployHost` is wrong.
    * If `deployAccount` is wrong.
* `CredentialNotFoundException`:
    * If the credentials cannot be resolved.

## Example
```groovy
neoDeploy script: this, archivePath: 'path/to/archiveFile.mtar', credentialsId: 'my-credentials-id'
```
