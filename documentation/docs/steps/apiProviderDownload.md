# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
apiProviderDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  apiProviderDownload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    apiProviderName: 'MY_API_PROVIDER_NAME'
    downloadPath: MY_API_PROVIDER_JSON_FILE_DOWNLOAD_PATH
```
