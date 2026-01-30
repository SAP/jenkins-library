# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* You have a user for the SAP BTP Cloud Foundry environment
* Credentials have been configured in Jenkins with a dedicated Id

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

Such a `manifest.yml` file needs to have the following structure, e.g. for creating three mongoDB Services with the Service Plan v4.0-dev:

```yaml

---
create-services:
- name:   "testDatabase1"
  broker: "mongodb"
  plan:   "v4.0-dev"

- name:   "testDatabase2"
  broker: "mongodb"
  plan:   "v4.0-dev"

- name:   "testDatabase3"
  broker: "mongodb"
  plan:   "v4.0-dev"
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

Additionally the Cloud Foundry Create-Service-Push Plugin offers the option to make use of variable substitution. This enables you to rename variables in the `manifest.yml` dynamically. It can be done either via providing the file path to a dedicated YAML file containing the information regarding the variable  substitution values in the `manifestVariablesFiles` flag or via providing a String List in the `manifestVariables` flag. Either ways can be achieved as seen in below examples for creating MongoDB instances.

For both ways you need to adapt the `manifest.yml` file to be relevant for variable substitution. This can be done according to below example:

```yaml

---
create-services:
- name:   ((name1))
  broker: "mongodb"
  plan:   "v4.0-dev"

- name:   ((name2))
  broker: "mongodb"
  plan:   "v4.0-dev"

- name:   ((name3))
  broker: "mongodb"
  plan:   "v4.0-dev"
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
