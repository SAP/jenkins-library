# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
scriptCollectionDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  scriptCollectionDownload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    scriptCollectionId: 'MY_SCRIPT_COLLECTION_NAME'
    scriptCollectionVersion: 'MY_SCRIPT_COLLECTION_VERSION'
    downloadPath: MY_SCRIPT_COLLECTION_DOWNLOAD_PATH
```
