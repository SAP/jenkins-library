# ${docGenStepName}

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.

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
