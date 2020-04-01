# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* This step is for creating a Service Key for an existing Service in Cloud Foundry.
* Cloud Foundry API endpoint, organization, space, user and service instance are available
* Credentials have been configured in Jenkins with a dedicated Id
* Additionally you can set the optional serviceKeyConfig flag to configure the Service Key creation with your respective JSON configuration.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

The following example creates a service key named "myServiceKey" for the service instance "myServiceInstance" in the provided cloud foundry organization and space. For the service key creation, the serviceKeyConfig is used.

```groovy
cloudFoundryCreateServiceKey(
  script: this,
  cloudFoundry: [
      apiEndpoint: 'https://test.server.com',
      credentialsId: 'cfCredentialsId',
      org: 'cfOrg',
      space: 'cfSpace',
      serviceInstance: 'myServiceInstance',
      serviceKeyName: 'myServiceKey',
      serviceKeyConfig: '{ \"key\" : \"value\" }'
  ])
```
