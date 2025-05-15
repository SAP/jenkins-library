# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

- Please make sure, that you are under Embedded Steampunk environment.
- Please make sure, that the service landscape-portal-api-for-s4hc with plan api was assigned as entitlement to the subaccount, where you are about to deploy addon product.
- Please make sure, that before deploying addon product, an instance of landscape-portal-api-for-s4hc (plan api) was created, and a service key with x509 authentication mechanism was created for the instance. The service key needs to be stored in the Jenkins Credentials Store.
- Please make sure, that the system to deploy addon product is active, and the descriptor file with deployment information is available.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example: Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapLandscapePortalUpdateAddOnProduct script: this
```

The configuration values for the addon update can be passed through the `config.yml` file:

```yaml
steps:
  abapLandscapePortalUpdateAddOnProduct:
    landscapePortalAPICredentialsId: 'landscapePortalAPICredentialsId'
    abapSystemNumber: 'abapSystemNumber'
    addonDescriptorFileName: 'addon.yml'
    addonDescriptor: 'addonDescriptor'
```

## Example: Configuration in the Jenkinsfile

The step, including all parameters, can also be called directly from the Jenkinsfile. In the following example, a configuration file is used.

```groovy
abapLandscapePortalUpdateAddOnProduct (
  script: this,
  landscapePortalAPICredentialsId: 'landscapePortalAPICredentialsId'
  abapSystemNumber: 'abapSystemNumber'
  addonDescriptorFileName: 'addon.yml'
  addonDescriptor: 'addonDescriptor'
)
```

The file `addon.yml` would look like this:

```yaml
addonProduct: some-addon-product
addonVersion: some-addon-version
```
