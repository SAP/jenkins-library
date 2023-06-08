# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
valueMappingArtifactDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  valueMappingArtifactDownload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    valueMappingId: 'MY_VALUE_MAPPING_NAME'
    valueMappingVersion: 'MY_VALUE_MAPPING_VERSION'
    downloadPath: MY_VALUE_MAPPING_PATH
```
