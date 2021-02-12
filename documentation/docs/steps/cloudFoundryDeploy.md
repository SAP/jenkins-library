# ${docGenStepName}

## ${docGenDescription}

### Additional hints

* Via parameter `useGoStep` it can be switched between
the groovy and the go implementation of that step. E.g. in case there are
issue with the go step it can be swtiched back to the corresponding groovy
code via `useGoStep:false` in the step configuration.

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
