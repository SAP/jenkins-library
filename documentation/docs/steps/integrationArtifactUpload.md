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
integrationArtifactUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactUpload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'MY_INTEGRATION_FLOW_ID'
    integrationFlowName: 'MY_INTEGRATION_FLOW_Name'
    packageId: 'MY_INTEGRATION_Package_ID'
    filePath: 'MY_INTEGRATION_FLOW_Artifact_Relative_Path'
    downloadPath: /MY_INTEGRATION_FLOW_DOWNLOAD_PATH
```
