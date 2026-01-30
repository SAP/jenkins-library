# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactDownload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'MY_INTEGRATION_FLOW_NAME'
    integrationFlowVersion: 'MY_INTEGRATION_FLOW_VERSION'
    downloadPath: MY_INTEGRATION_FLOW_DOWNLOAD_PATH
```
