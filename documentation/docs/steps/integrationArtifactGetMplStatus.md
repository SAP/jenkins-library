# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactGetMplStatus script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactGetMplStatus:
    cpiCredentialsId: 'MY_CPI_OAUTH_CREDENTIALSID_IN_JENKINS'
    integrationFlowId: 'INTEGRATION_FLOW_ID'
    platform: cf
    host: https://CPI_HOST_ITSPACES_URL
    oAuthTokenProviderUrl: https://CPI_HOST_OAUTH_URL
```
