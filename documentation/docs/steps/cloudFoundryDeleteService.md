# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step is for deleting an exisiting service on Cloud Foundry.
You need to provide the Cloud Foundry API Endpoint, the Organisation as well as the space and the respective Service instance name you want to delete. 
Furhtermore you will need to provide the Cloud Foudnry Login credentials, which must be stored in the Jenkins configuration.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

In this example, the Cloud Foundry configuration is directly provided with the respective credentials for the used user/account.

```groovy
cloudFoundryDeleteService(
    cfapiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    space: 'cfSpace',
    serviceInstance: 'cfServiceInstance',
    cfCredentialsId: 'cfCredentialsId',
) 
```