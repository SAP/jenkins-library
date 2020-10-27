# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have a user for the SAP Cloud Platform Cloud Foundry Environment
* Credentials have been configured in Jenkins with a dedicated Id
* An instance of Extension Factory Serverless Runtime (xfsrt) running in the target cf space
* Service key created for xfsrt service instance

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

1. Deploying **without** deployment configuration data:

    ```groovy
    cloudFoundryFaasDeploy(
        cfCredentialsId: "<CF_LOGIN_CREDENTIAL>"
        cfApiEndpoint: "<CF_API_ENDPOINT>"
        cfOrg: "<CF_ORG>"
        cfSpace: "<CF_SPACE>"
        xfsrtServiceInstance: "<XFSRT_SERVICE_INSTANCE_NAME>"
        xfsrtServiceKeyName: "<SERVICE-KEY>" //service key created for xfsrt instance
    )
    ```

1. Deploying **with** deployment configuration data:

    ```groovy
    cloudFoundryFaasDeploy(
        cfCredentialsId: "<CF_LOGIN_CREDENTIAL>"
        cfApiEndpoint: "<CF_API_ENDPOINT>"
        cfOrg: "<CF_ORG>"
        cfSpace: "<CF_SPACE>"
        xfsrtServiceInstance: "<XFSRT_SERVICE_INSTANCE_NAME>"
        xfsrtServiceKeyName: "<SERVICE-KEY>" //service key created for xfsrt instance
        xfsrtValuesCredentialsId: "<SECRET_TEXT_CREDENTIAL_ID>" //the id of a secret text credential, which contains a json string required during the deployment
    )
    ```

    Using the `xfsrt-cli` one can easily generate an initial deployment values json string based on specific secret definitions. To do so, run inside the project:

    ```bash
    xfsrt-cli faas project init values -o json
    ```

   Then the initial dummy values have to be changed to real values, and the json string added to a secret text credential.
