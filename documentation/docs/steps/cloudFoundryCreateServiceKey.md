# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* This step is for creating a Service Key for an existing Service in Cloud Foundry.
* Cloud Foundry API endpoint, Organization, Space, user and Service Instance are available
* Credentials have been configured in Jenkins with a dedicated Id
* Additionally you can set the optional `serviceKeyConfig` flag to configure the Service Key creation with your respective JSON configuration. The JSON configuration can either be a JSON or the path a dedicated JSON configuration file containing the JSON configuration. If you chose a dedicated config file, it must be stored in a file that must be referenced in the `serviceKeyConfigFile` flag. You must store the file in the same folder as your `Jenkinsfile` that starts the Pipeline in order for the Pipeline to be able to find the file. Most favourable SCM is Git.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

The following examples will create a Service Key named "myServiceKey" for the Service Instance "myServiceInstance" in the provided Cloud Foundry Organization and Space. For the Service Key creation in these example, the serviceKeyConfig parameter is used. It will show the different ways of passing the JSON configuration, either via a string or the path to a file containing the JSON configuration.
If you dont want to use a special configuration simply remove the parameter since it is optional.

### Create Service Key with JSON config file in Jenkinsfile

This example covers the parameters for a Jenkinsfile when using the cloudFoundryCreateServiceKey step. It uses a `serviceKeaConfig.json` file with valid JSON objects for creating a Cloud Foundry Service Key.

```groovy
cloudFoundryCreateServiceKey(
  cfApiEndpoint: 'https://test.server.com',
  cfCredentialsId: 'cfCredentialsId',
  cfOrg: 'cfOrg',
  cfSpace: 'cfSpace',
  cfServiceInstance: 'myServiceInstance',
  cfServiceKeyName: 'myServiceKey',
  cfServiceKeyConfig: 'serviceKeyConfig.json',
  script: this,
)
```

The JSON config file, e.g. `serviceKeyConfig.json` can look like this:

```json
{
  "example":"value",
  "example":"value"
}
```

### Create Service Key with JSON string in Jenkinsfile

The following example covers the creation of a Cloud Foundry Service Key in a Jenkinsfile with using a JSON string as a config for the Service Key creation. If you use a Jenkinsfile for passing the parameter values you need to escape the double quotes in the JSON config string.

```groovy
cloudFoundryCreateServiceKey(
  cfApiEndpoint: 'https://test.server.com',
  cfCredentialsId: 'cfCredentialsId',
  cfOrg: 'cfOrg',
  cfSpace: 'cfSpace',
  cfServiceInstance: 'myServiceInstance',
  cfServiceKeyName: 'myServiceKey',
  cfServiceKeyConfig: '{\"example\":\"value\",\"example\":\"value\"}',
  script: this,
)
```

### Create Service Key with JSON string as parameter in .pipeline/config.yml file

If you chose to provide a `config.yml` file you can provide the parameters including the values in this file. You only need to set the script parameter when calling the step:

```groovy
cloudFoundryCreateServiceKey(
  script: this,
)
```

The `.pipeline/config.yml` has to contain the following parameters accordingly:

```yaml
steps:
    cloudFoundryCreateServiceKey:
        cfApiEndpoint: 'https://test.server.com'
        cfOrg: 'testOrg'
        cfSpace: 'testSpace'
        cfServiceInstance: 'testInstance'
        cfServiceKeyName: 'myServiceKey'
        cfServiceKeyConfig: '{"example":"value","example":"value"}'
        cfCredentialsId: 'cfCredentialsId'
```

When using a `.pipeline/config.yml` file you don't need to escape the double quotes in the JSON config string.
You can also pass the path to a JSON config file in the `cfServiceKeyConfig` parameter. Example: `cfServiceKeyConfig: 'serviceKeyconfig.json'`
