# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have a user for the SAP Cloud Platform Cloud Foundry Environment
* Credentials have been configured in Jenkins with a dedicated Id
* An instance of Extension Factory Serverless Runtime(xfsrt)running in the target cf space
* Service key created for xfsrt service instance

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

* ### Deploying functions into xfsrt environment without deplyoment configuration data

The following example deploys the functions project into xfsrt instance running in the cloud foundry space.


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

* ### Deploying functions into xfsrt environment with deployment configuration data

```groovy
cloudFoundryFaasDeploy(
    cfCredentialsId: "<CF_LOGIN_CREDENTIAL>"
    cfApiEndpoint: "<CF_API_ENDPOINT>"
    cfOrg: "<CF_ORG>"
    cfSpace: "<CF_SPACE>"
    xfsrtServiceInstance: "<XFSRT_SERVICE_INSTANCE_NAME>"
    xfsrtServiceKeyName: "<SERVICE-KEY>" //service key created for xfsrt instance
    xfsrtValuesCredentialsId: "<SECRET_TEXT_CREDENTIAL_ID>" //secret text credential containing a json string(secret credential) required during the deployment.
)
```

