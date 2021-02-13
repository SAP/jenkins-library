# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactUpload:
    cpiCredentialsId: 'MY_CPI_OAUTH_CREDENTIALSID_IN_JENKINS'
    integrationFlowId: 'MY_INTEGRATION_FLOW_ID'
    integrationFlowVersion: 'MY_INTEGRATION_FLOW_VERSION'
    integrationFlowName: 'MY_INTEGRATION_FLOW_Name'
    packageId: 'MY_INTEGRATION_Package_ID'
    filePath: 'MY_INTEGRATION_FLOW_Artifact_Relative_Path'
    host: https://CPI_HOST_ITSPACES_URL
    oAuthTokenProviderUrl: https://CPI_HOST_OAUTH_URL
    downloadPath: /MY_INTEGRATION_FLOW_DOWNLOAD_PATH
```
