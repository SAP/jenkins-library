# ${docGenStepName}

## ${docGenDescription}

API Provider defines the connection details for services running on specific hosts. Apart from the details of the specific hosts, API provider can also be used to define any further details that are necessary to establish the connection. This function returns the APIProvider entity and store it in the file system by providing the name.

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
