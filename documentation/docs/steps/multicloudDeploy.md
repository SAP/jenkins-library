# ${docGenStepName}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

```groovy
multicloudDeploy(
    script: script,
    cfTargets: [[apiEndpoint: 'https://test.server.com', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', org: 'cfOrg', space: 'cfSpace']],
    neoTargets: [[credentialsId: 'my-credentials-id', host: hana.example.org, account: 'trialuser1']],
    enableZeroDowntimeDeployment: 'true'
)
```
