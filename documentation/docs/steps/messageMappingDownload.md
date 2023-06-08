# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
messageMappingDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  messageMappingDownload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    messageMappingId: 'MY_MESSAGE_MAPPING_NAME'
    messageMappingVersion: 'MY_MESSAGE_MAPPING_VERSION'
    downloadPath: MY_MESSAGE_MAPPING_DOWNLOAD_PATH
```
