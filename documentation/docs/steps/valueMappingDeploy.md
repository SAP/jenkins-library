# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
valueMappingDeploy script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  valueMappingDeploy:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    valueMappingId: 'MY_VALUE_MAPPING_NAME'
```
