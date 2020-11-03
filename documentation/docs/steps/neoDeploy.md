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

## Example for invalidating the cache

Setting parameter `invalidateCache` to `true`, will clean up the cache of a Fiori Launchpad site, refreshing the content of html5 applications deployed there. This is **applicable** only when the application (**only** html5 applications ) is deployed, accessed via Portal service (Fiori Launchpad site).

Setting this parameter to true, needs additional configurations to go through. Firstly, create an OAuth credential as shown below:

### OAuth credential creation

1. Select the OAuth settings from your subaccount, create a new client with subscription to portal/nwc as shown in the following images:

    ![OAuth client creation](../images/oauthClientCreation.png)

2. Select the "Clients" tab, which provides an option to register a new client. Click on "Register New Client" button.

3. Then, in the subscription field, select the portal landscape you would like to subscribe to, ex: `portal/nwc` or `portal/sandbox` as shown below:

    ![Portal subscription](../images/portalSubscription.png)

4. In the "Authorization Grant" field, select "Client Credentials" from the drop down menu. Then, enter the "Secret" field with a user defined password and finally save the changes.

After saving these changes, create a UsernamePassword type credential with client Id (obtained from registering new client) as username and client secret (password entered in the secret field above) as password in Jenkins.

### Set Site Id

1. After login to the portal service from SAP CP, the user has a possibility to create a site directory, within which the deployed application resides.

2. On the site directory tile,  click on "edit" button, which lists some set of option as shown in the below image. By publishing that site, the control is navigated to a new page which holds the "site id" in the url. By default, site is set as default as shown below image.

    ![Site set to default](../images/portalSiteSetToDefault.png)

3. If it is already set to default, no need to configure `siteId` in the configuration file. If not set to default, please configure `siteId`, as shown in the below configuration:

Example configuration:

```yaml
steps:
  <...>
  neoDeploy:
    neo:
      account: <myDeployAccount>
      host: hana.example.org
      credentialsId: 'my-credentials-id'
      invalidateCache: true
      portalLandscape: "cloudnwcportal"
      oauthCredentialId: <OAUTH_CREDENTIAL_ID>
      siteId: <PORTAL_SITE_ID> # not required, if the default site is already set in the portal service (SAP CP)
```
