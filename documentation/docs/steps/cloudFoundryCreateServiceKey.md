# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* Cloud Foundry API endpoint, organization, space, user and service instance are available
* Credentials have been configured in Jenkins with a dedicated Id

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
      serviceKey: 'myServiceKey',
      serviceKeyConfig: '{ \"key\" : \"value\" }'
  ])
```
