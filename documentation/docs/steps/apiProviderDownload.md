# ${docGenStepName}

## ${docGenDescription}

* API Provider defines the connection details for services running on specific hosts.
* This function returns the APIProvider entity and stores it in the file system.

## Prerequisites

API Provider artifact to be downloaded should exist in the API Portal.

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
