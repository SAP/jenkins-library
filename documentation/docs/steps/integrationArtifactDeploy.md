# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactDeploy script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactDeploy:
    cpiCredentialsId: 'MY_CPI_OAUTH_CREDENTIALSID_IN_JENKINS'
    integrationFlowId: 'MY_INTEGRATION_FLOW_NAME'
    integrationFlowVersion: 'MY_INTEGRATION_FLOW_VERSION'
    platform: cf
    host: https://CPI_HOST_ITSPACES_URL
    oAuthTokenProviderUrl: https://CPI_HOST_OAUTH_URL
```
