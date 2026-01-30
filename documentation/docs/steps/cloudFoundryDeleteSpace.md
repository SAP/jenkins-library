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

### Space deletion in Cloud Foundry with a simple example

The following example deletes an existing space in a Cloud Foundry.

You can store the credentials in Jenkins and use the `cfCredentialsId` parameter to authenticate to Cloud Foundry.

This can be done accordingly:

```groovy
cloudFoundryDeleteSpace(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace', //Name of the cf space to be deleted
    cfCredentialsId: 'cfCredentialsId',
    script: this,
)
```
