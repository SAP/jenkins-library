# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step is for deleting an exisiting service on Cloud Foundry

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

In this example, the Cloud Foundry configuration is directly provided with the respective credentials for the used user/account.

```groovy
cloudFoundryDeleteService(
                    cloudFoundry : [
                        apiEndpoint : 'https://test.server.com',
                        org : 'cfOrg',
                        space: 'cfSpace',
                        serviceInstance: 'cfServiceInstance',
                    ],
                    cfCredentialsId: 'cfCredentialsId',
                ) 
```