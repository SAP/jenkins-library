# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
apiProxyUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  apiProxyUpload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    filePath: MY_API_PROXY_ZIP_FILE_PATH
```
