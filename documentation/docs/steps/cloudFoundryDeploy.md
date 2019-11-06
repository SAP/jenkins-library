# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* Cloud Foundry organization, space and deployment user are available
* Credentials for deployment have been configured in Jenkins with a dedicated Id

    ![Jenkins credentials configuration](../images/cf_credentials.png)

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
cloudFoundryDeploy(
    script: script,
    deployType: 'blue-green',
    cloudFoundry: [apiEndpoint: 'https://test.server.com', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', org: 'cfOrg', space: 'cfSpace'],
    deployTool: 'cf_native'
)
```
