# neoDeploy

## Description
Deploys an Application to SAP Cloud Platform (SAP CP) using the SAP Cloud Platform Console Client (Neo Java Web SDK).

Before doing this, validates that SAP Cloud Platform Console Client is installed and the version is compatible.

Note that a version is formed by `major.minor.patch`, and a version is compatible to another version if the minor and patch versions are higher, but the major version is not, e.g. if 3.39.10 is the expected version, 3.39.11 and 3.40.1 would be compatible versions, but 4.0.1 would not be a compatible version.

## Prerequisites
* **SAP CP account** - the account to where the application is deployed.
* **SAP CP user for deployment** - a user with deployment permissions in the given account.
* **Jenkins credentials for deployment** - must be configured in Jenkins credentials with a dedicated Id.

![Jenkins credentials configuration](../images/neo_credentials.png)

* **Neo Java Web SDK 3.39.10 or compatible version** - can be downloaded from [Maven Central](http://central.maven.org/maven2/com/sap/cloud/neo-java-web-sdk/). The Neo Java Web SDK
needs to be extracted into the folder provided by `neoHome`. In case this parameters is not provided and there is no NEO_HOME parameter in the environment
`<neoRoot>/tools` needs to be in the `PATH`. This step is also capable of triggering the neo deploy tool provided inside a docker image.

* **Java 8 or compatible version** - needed by the *Neo-Java-Web-SDK*

## Parameters when using MTA deployment method (default - MTA)
| parameter          | mandatory | default                       | possible values                                 |
| -------------------|-----------|-------------------------------|-------------------------------------------------|
| `account`          | no        |                               |                                                 |
| `archivePath`      | yes       |                               |                                                 |
| `deployAccount` deprecated, use `account` | no  |              |                                                 |
| `deployHost` deprecated, use `host` | no        |              |                                                 |
| `deployMode`       | yes       | `'mta'`                       | `'mta'`, `'warParams'`, `'warPropertiesFile'`   |
| `host`             | no        |                               |                                                 |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`         |                                                 |
| `neoHome`          | no        |                               |                                                 |
| `script`           | yes       |                               |                                                 |

## Parameters when using WAR file deployment method with .properties file (WAR_PROPERTIESFILE)
| parameter          | mandatory | default                       | possible values                                 |
| -------------------|-----------|-------------------------------|-------------------------------------------------|
| `archivePath`      | yes       |                               |                                                 |
| `deployMode`       | yes       | `'mta'`                       | `'mta'`, `'warParams'`, `'warPropertiesFile'`   |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`         |                                                 |
| `neoHome`          | no        |                               |                                                 |
| `propertiesFile`   | yes       |                               |                                                 |
| `script`           | yes       |                               |                                                 |
| `warAction`        | yes       | `'deploy'`                    | `'deploy'`, `'rolling-update'`                  |

## Parameters when using WAR file deployment method witout .properties file - with parameters (WAR_PARAMS)
| parameter          | mandatory | default                       | possible values                                 |
| -------------------|-----------|-------------------------------|-------------------------------------------------|
| `account`          | no        |                               |                                                 |
| `applicationName`  | yes       |                               |                                                 |
| `archivePath`      | yes       |                               |                                                 |
| `deployAccount` deprecated, use `account` | no                 |                                                 |
| `deployHost` deprecated, use `host` | no                       |                                                 |
| `deployMode`       | yes       | `'mta'`                       |  `'mta'`, `'warParams'`, `'warPropertiesFile'`  |
| `host`             | no        |                               |                                                 |
| `neoCredentialsId` | no        | `'CI_CREDENTIALS_ID'`         |                                                 |
| `neoHome`          | no        |                               |                                                 |
| `runtime`          | yes       |                               |                                                 |
| `runtime-version`  | yes       |                               |                                                 |
| `script`           | yes       |                               |                                                 |
| `vmSize`           | no        | `'lite'`                      | `'lite'`, `'pro'`, `'prem'`, `'prem-plus'`      |
| `warAction`        | yes       | `'deploy'`                    | `'deploy'`, `'rolling-update'`                  |


* `account` - The SAP Cloud Platform account to deploy to.
* `applicationName` - Name of the application you want to manage, configure, or deploy
* `archivePath`- The path to the archive for deployment to SAP CP.
* `deployAccount` - deprecated, use `account`. The SAP Cloud Platform account to deploy to.
* `deployHost` - deprecated, use `host`. The SAP Cloud Platform host to deploy to.
* `deployMode` - The deployment mode which should be used. Available options are `'mta'` (default), `'warParams'` (deploying WAR file and passing all the deployment parameters via the function call) and `'warPropertiesFile'` (deploying WAR file and putting all the deployment parameters in a .properties file)
* `host` - The SAP Cloud Platform host to deploy to.
* `neoCredentialsId` - The Jenkins credentials containing user and password used for SAP CP deployment.
* `neoHome` - The path to the `neo-java-web-sdk` tool used for SAP CP deployment. If no parameter is provided, the path is retrieved from the environment variables using the environment variable `NEO_HOME`. If no parameter and no environment variable is provided, the path is retrieved from the step configuration using the step configuration key `neoHome`. If the previous configurations are not provided, the tool is expected on the `PATH`, and if it is not available on the `PATH` an AbortException is thrown.
* `propertiesFile` - The path to the .properties file in which all necessary deployment properties for the application are defined.
* `runtime` - Name of SAP Cloud Platform application runtime
* `runtime-version` - Version of SAP Cloud Platform application runtime
* `script` - The common script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `vmSize` - Compute unit (VM) size. Acceptable values: lite, pro, prem, prem-plus.
* `warAction` - Action mode when using WAR file mode. Available options are `deploy` (default) and `rolling-update` which performs update of an application without downtime in one go.

The step is prepared for being executed in docker. The corresponding parameters can be applied. See step `dockerExecute` for details.

## Step configuration
The following parameters can also be specified as step parameters using the global configuration file:

* `account`
* `dockerEnvVars`
* `dockerImage`
* `dockerOptions`
* `host`
* `neoCredentialsId`
* `neoHome`

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

Example configuration:

```
steps:
  <...>
  neoDeploy:

        account: <myDeployAccount>
        host: hana.example.org
```
