# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
apiKeyValueMapUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  apiKeyValueMapUpload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    key: API_KEY_NAME
    value: API_KEY_VALUE
    keyValueMapName: API_KEY_VALUE_MAP_NAME
```
