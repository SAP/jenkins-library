# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
messageMappingUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  messageMappingUpload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    messageMappingId: 'MY_MESSAGE_MAPPING_ID'
    messageMappingName: 'MY_MESSAGE_MAPPING_Name'
    packageId: 'MY_INTEGRATION_Package_ID'
    filePath: 'MY_MESSAGE_MAPPING_Artifact_Relative_Path'
    downloadPath: /MY_MESSAGE_MAPPING_DOWNLOAD_PATH
```
