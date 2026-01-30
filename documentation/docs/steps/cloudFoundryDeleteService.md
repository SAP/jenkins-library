# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

This step is for deleting an existing service on Cloud Foundry.
You need to provide the Cloud Foundry API Endpoint, the Organisation as well as the Space and the respective Service Instance Name you want to delete.
Furthermore you will need to provide the Cloud Foundry Login Credentials, which must be stored in the Jenkins Configuration.
Additionally you can set the cfDeleteServiceKeys flag for deleting all Service Keys that belong to the respective Service.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

In this example, the Cloud Foundry Configuration is directly provided with the respective Credentials for the used User/Account.

```groovy
cloudFoundryDeleteService(
    cfApiEndpoint: 'https://test.server.com',
    cfOrg: 'cforg',
    cfSpace: 'cfspace',
    cfServiceInstance: 'cfserviceInstance',
    cfCredentialsId: 'cfcredentialsId',
    cfDeleteServiceKeys: true,
)
```
