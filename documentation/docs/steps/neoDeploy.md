# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* **SAP CP account** - the account to where the application is deployed. To deploy MTA (`deployMode: mta`) an over existing _Java_ application, free _Java Quota_ of at least 1 is required, which means that this will not work on trial accounts.
* **SAP CP user for deployment** - a user with deployment permissions in the given account.
* **Jenkins credentials for deployment** - must be configured in Jenkins credentials with a dedicated Id.

![Jenkins credentials configuration](../images/neo_credentials.png)

* **Neo Java Web SDK 3.39.10 or compatible version** - can be downloaded from [Maven Central](http://central.maven.org/maven2/com/sap/cloud/neo-java-web-sdk/). This step is capable of triggering the neo deploy tool provided inside a docker image. We provide docker image `ppiper/neo-cli`. `neo.sh` needs to be contained in path, e.g by adding a symbolic link to `/usr/local/bin`.

* **Java 8 or compatible version** - needed by the *Neo-Java-Web-SDK*. Java environment needs to be properly configured (JAVA_HOME, java exectutable contained in path).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Exceptions

* `Exception`:
    * If `source` is not provided.
    * If `propertiesFile` is not provided (when using `'WAR_PROPERTIESFILE'` deployment mode).
    * If `application` is not provided (when using `'WAR_PARAMS'` deployment mode).
    * If `runtime` is not provided (when using `'WAR_PARAMS'` deployment mode).
    * If `runtimeVersion` is not provided (when using `'WAR_PARAMS'` deployment mode).
* `AbortException`:
    * If neo-java-web-sdk is not properly installed.
* `CredentialNotFoundException`:
    * If the credentials cannot be resolved.

## Example

```groovy
neoDeploy script: this, source: 'path/to/archiveFile.mtar', neo: [credentialsId: 'my-credentials-id', host: hana.example.org]
```

Example configuration:

```yaml
steps:
  <...>
  neoDeploy:
    deployMode: mta
    neo:
      account: <myDeployAccount>
      host: hana.example.org
```
