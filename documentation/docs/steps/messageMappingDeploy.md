# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
messageMappingDeploy script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  messageMappingDeploy:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    messageMappingId: 'MY_MESSAGE_MAPPING_NAME'
```
