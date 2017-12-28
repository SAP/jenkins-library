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

## Parameters when using MTA deployment method (default - MTA)
| parameter          | mandatory | default                                                                                      | possible values                                 |
| -------------------|-----------|----------------------------------------------------------------------------------------------|-------------------------------------------------|
| `deployMode`       | yes       | `'MTA'`                                                                                      | `'MTA'`, `'WAR_PARAMS'`, `'WAR_PROPERTIESFILE'` |
| `script`           | yes       |                                                                                              |                                                 |
| `archivePath`      | yes       |                                                                                              |                                                 |
| `deployHost`       | no        | `'DEPLOY_HOST'` from `commonPipelineEnvironment`                                             |                                                 |
| `deployAccount`    | no        | `'CI_DEPLOY_ACCOUNT'` from `commonPipelineEnvironment`                                       |                                                 |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`                                                                        |                                                 |
| `neoHome`          | no        | Environment is checked for `NEO_HOME`, <br>otherwise the neo toolset is expected in the path |                                                 |

## Parameters when using WAR file deployment method with .properties file (WAR_PROPERTIESFILE)
| parameter          | mandatory | default                                                                                      | possible values                                 |
| -------------------|-----------|----------------------------------------------------------------------------------------------|-------------------------------------------------|
| `deployMode`       | yes       | `'MTA'`                                                                                      | `'MTA'`, `'WAR_PARAMS'`, `'WAR_PROPERTIESFILE'` |
| `warAction`        | yes       | `'deploy'`                                                                                   | `'deploy'`, `'rolling-update'`                  |
| `script`           | yes       |                                                                                              |                                                 |
| `archivePath`      | yes       |                                                                                              |                                                 |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`                                                                        |                                                 |
| `neoHome`          | no        | Environment is checked for `NEO_HOME`, <br>otherwise the neo toolset is expected in the path |                                                 |
| `propertiesFile`   | yes       |                                                                                              |                                                 |

## Parameters when using WAR file deployment method witout .properties file - with parameters (WAR_PARAMS)
| parameter          | mandatory | default                                                                                      | possible values                                 |
| -------------------|-----------|----------------------------------------------------------------------------------------------|-------------------------------------------------|
| `deployMode`       | yes       | `'MTA'`                                                                                      | `'MTA'`, `'WAR_PARAMS'`, `'WAR_PROPERTIESFILE'` |
| `warAction`        | yes       | `'deploy'`                                                                                   | `'deploy'`, `'rolling-update'`                  |
| `script`           | yes       |                                                                                              |                                                 |
| `archivePath`      | yes       |                                                                                              |                                                 |
| `deployHost`       | no        | `'DEPLOY_HOST'` from `commonPipelineEnvironment`                                             |                                                 |
| `deployAccount`    | no        | `'CI_DEPLOY_ACCOUNT'` from `commonPipelineEnvironment`                                       |                                                 |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`                                                                        |                                                 |
| `neoHome`          | no        | Environment is checked for `NEO_HOME`, <br>otherwise the neo toolset is expected in the path |                                                 |
| `applicationName`  | yes       |                                                                                              |                                                 |
| `runtime`          | yes       |                                                                                              |                                                 |
| `runtime-version`  | yes       |                                                                                              |                                                 |
| `size`             | no        | `'lite'`                                                                                     | `'lite'`, `'pro'`, `'prem'`, `'prem-plus'`      |


* `deployMode` - The deployment mode which should be used. Available options are `'MTA'` (default), `'WAR_PARAMS'` (deploying WAR file and passing all the deployment parameters via the function call) and `'WAR_PROPERTIESFILE'` (deploying WAR file and putting all the deployment parameters in a .properties file)
* `script` - The common script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `archivePath`- The path to the archive for deployment to SAP CP.
* `deployHost` - The SAP Cloud Platform host to deploy to.
* `deployAccount` - The SAP Cloud Platform account to deploy to.
* `credentialsId` - The Jenkins credentials containing user and password used for SAP CP deployment.
* `neoHome` - The path to the `neo-java-web-sdk` tool used for SAP CP deployment. If no parameter is provided, the path is retrieved from the Jenkins environment variables using `env.NEO_HOME`. If this Jenkins environment variable is not set it is assumed that the tool is available in the `PATH`.
* `propertiesFile` - The path to the .properties file in which all necessary deployment properties for the application are defined.
* `warAction` - Action mode when using WAR file mode. Available options are `deploy` (default) and `rolling-update` which performs update of an application without downtime in one go.
* `applicationName` - Name of the application you want to manage, configure, or deploy
* `runtime` - Name of SAP Cloud Platform application runtime
* `runtime-version` - Version of SAP Cloud Platform application runtime
* `size` - Compute unit (VM) size. Acceptable values: lite, pro, prem, prem-plus.

## Return value
none

## Side effects
none

## Exceptions
* `Exception`:
    * If `archivePath` is not provided.
    * If `propertiesFile` is not provided (when using `'WAR_PROPERTIESFILE'` deployment mode).
    * If `applicationName` is not provided (when using `'WAR_PARAMS'` deployment mode).
    * If `runtime` is not provided (when using `'WAR_PARAMS'` deployment mode).
    * If `runtime-version` is not provided (when using `'WAR_PARAMS'` deployment mode).
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
