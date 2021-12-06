# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
apiKeyValueMapDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  apiKeyValueMapDownload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    apiProxyName: 'MY_API_KEY_VALUE_MAP_NAME'
    downloadPath: MY_API_KEY_VALUE_MAP_CSV_FILE_DOWNLOAD_PATH
```
