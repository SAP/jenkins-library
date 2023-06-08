# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
scriptCollectionUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  scriptCollectionUpload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    scriptCollectionId: 'MY_SCRIPT_COLLECTION_ID'
    scriptCollectionName: 'MY_SCRIPT_COLLECTION_Name'
    packageId: 'MY_INTEGRATION_Package_ID'
    filePath: 'MY_SCRIPT_COLLECTION_Artifact_Relative_Path'
    downloadPath: /MY_SCRIPT_COLLECTION_DOWNLOAD_PATH
```
