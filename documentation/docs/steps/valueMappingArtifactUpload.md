# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
valueMappingArtifactUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  valueMappingArtifactUpload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    valueMappingId: 'MY_VALUE_MAPPING_ID'
    valueMappingName: 'MY_VALUE_MAPPING_Name'
    packageId: 'MY_INTEGRATION_Package_ID'
    filePath: 'MY_VALUE_MAPPING_Artifact_Relative_Path'
    downloadPath: /MY_VALUE_MAPPING_DOWNLOAD_PATH
```
