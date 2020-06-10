# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* This step is for creating one or multiple Services in Cloud Foundry.
* Cloud Foundry API endpoint, Organization, Space and user are available
* If you chose to create a single Service the Service Instance Name, Plan Broker of the Service to be created have to be available
* Credentials have been configured in Jenkins with a dedicated Id
* You can set the optional `cfCreateServiceConfig` flag to configure the Service creation with your respective JSON configuration. The JSON configuration can either be an in-line JSON string or the path a dedicated JSON configuration file containing the JSON configuration. If you chose a dedicated config file, you must store the file in the same folder as your `Jenkinsfile` that starts the Pipeline in order for the Pipeline to be able to find the file. Most favourable SCM is Git.
* If you want the service to be created from a particular broker you can set the optional `cfServiceBroker`flag.
* You can set user provided tags for the Service creation using a flat list as the value for the optional `cfServiceTags` flag.
* Also you can create one or multiple Cloud Foundry Services at once with the Cloud Foundry Create-Service-Push Plugin using the optional `serviceManifest` flag. If you chose to set this flag, the Create-Service-Push Plugin will be used for all Service creations in this step and you will need to provide a `serviceManifest.yml` file. In that case, above described flags and options will not be used for the Service creations, since you chose to use the Create-Service-Push Plugin. Please see below examples for more information on how to make use of the plugin with the appropriate step configuation. Additionally the Plugin provides the option to make use of variable substitution for the Service creations. You can find further information regarding the functionality of the Cloud Foundry Create-Service-Push Plugin in the respective documentation: [Cloud Foundry Create-Service-Push Plugin](https://github.com/dawu415/CF-CLI-Create-Service-Push-Plugin)

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

* ### Single Service Creation in Cloud Foundry example with JSON-configuration in Jenkinsfile

The following example creates a single Service in Cloud Foundry. It makes use of the `cfCreateServiceConfig` flag for passing a JSON configuration as an in-line parameter string as well as the `cfServiceTags` for providing user tags.

You can store the credentials in Jenkins and use the `cfCredentialsId` parameter to authenticate to Cloud Foundry.

This can be done accordingly:

```groovy
cloudFoundryCreateService(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfCredentialsId: 'cfCredentialsId',
    cfService:  'myService',
    cfServiceInstanceName: 'myServiceInstanceName',
    cfServicePlan: 'myPlan',
    cfCreateServiceConfig: '{\"example\":\"value\",\"example\":\"value\"}',
    cfServiceTags: 'list, of, tags',
    script: this,
)
```

If you chose to having a dedicated JSON file for the JSON configuration for the `cfCreateServiceConfig` flag you can do so by referencing the file path accordingly. This file should be stored in the same folder as your `Jenkinsfile` that starts the Pipeline in order for the Pipeline to be able to find the file. Most favourable SCM is Git.
Such a JSON file with the appropriate step configuration could look as follows:

The JSON config file, e.g. `createServiceConfig.json` can look like this:

```json
{
  "example":"value",
  "example":"value"
}
```

The step configuration needs to contain the path to the JSON file:

```groovy
cloudFoundryCreateService(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfCredentialsId: 'cfCredentialsId',
    cfService:  'myService',
    cfServiceInstanceName: 'myServiceInstanceName',
    cfServicePlan: 'myPlan',
    cfCreateServiceConfig: 'createServiceConfig.json',
    cfServiceTags: 'list, of, tags',
    script: this,
)
```

* ### Multiple Service Creation in Cloud Foundry example with manifest file in Jenkinsfile

The following example shows the option to create multiple Services in Cloud Foundry. It makes use of the Cloud Foundry Create-Service-Push Plugin. This is described in above Prerequisites, please check this section for further information regarding its usage. This plugin enables this step to create multiple Cloud Foundry Services in one step.

It requires a dedicated YAML file, e.g. `manifest.yml`, that contains all the information for creating the services, including their names, service plan and the service broker.

Such a `manifest.yml` file needs to have the following structure:

```yaml

---
create-services:
- name:   "testService1"
  broker: "testBroker"
  plan:   "testPlan"

- name:   "testService2"
  broker: "testBroker"
  plan:   "testPlan"

- name:   "testService2"
  broker: "testBroker"
  plan:   "testPlan"
```

The path of the `manifest.yml` config file needs to be passed as a parameter in the `serviceManifest` flag.
You can store the credentials in Jenkins and use the `cfCredentialsId` parameter to authenticate to Cloud Foundry.

This can be done accordingly:

```groovy
cloudFoundryCreateService(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfCredentialsId: 'cfCredentialsId',
    serviceManifest: 'manifest.yml',
    script: this,
)
```

* ### Multiple Service Creation in Cloud Foundry example with manifest file and variable substitution in Jenkinsfile

Additionally the Cloud Foundry Create-Service-Push Plugin offers the option to make use of variable substitution. This enables you to rename variables in the `manifest.yml` dynamically. It can be done either via providing the file path to a dedicated YAML file containing the information regarding the variable  substitution values in the `manifestVariablesFiles` flag or via providing a String List in the `manifestVariables` flag. Either ways can be achieved as seen in below examples.

For both ways you need to adapt the `manifest.yml` file to be relevant for variable substitution. This can be done according to below example:

```yaml

---
create-services:
- name:   ((name1))
  broker: "testBroker"
  plan:   "testPlan"

- name:   ((name2))
  broker: "testBroker"
  plan:   "testPlan"

- name:   ((name3))
  broker: "testBroker"
  plan:   "testPlan"
```

If you chose to have a dedicated file for the variable substitution values, it needs to have the following structure of the `vars.yml` file:

```yaml
name1: test1
name2: test2
name3: test3
```

The path of the `manifest.yml` config file needs to be passed as a parameter in the `serviceManifest` flag as well as the path to the `vars.yml` file in the `manifestVariablesFiles` flag.
You can store the credentials in Jenkins and use the `cfCredentialsId` parameter to authenticate to Cloud Foundry.

This can be done accordingly:

```groovy
cloudFoundryCreateService(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfCredentialsId: 'cfCredentialsId',
    serviceManifest: 'manifest.yml',
    manifestVariablesFiles: 'vars.yml',
    script: this,
)
```

You can also pass the values for the variable substition as a string list for the `manifestVariables` flag. This needs to follow the pattern key=value.
This can be done accordingly:

```groovy
cloudFoundryCreateService(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfCredentialsId: 'cfCredentialsId',
    serviceManifest: 'manifest.yml',
    manifestVariables: ["name1=test1","name2=test2", "name3=test3"],
    script: this,
)
```
