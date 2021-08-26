# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactUpdateConfiguration script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactUpdateConfiguration:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'MY_INTEGRATION_FLOW_NAME'
    integrationFlowVersion: 'MY_INTEGRATION_FLOW_VERSION'
    parameterKey: 'MY_INTEGRATION_FLOW_CONFIG_PARAMETER_NAME'
    parameterValue: 'MY_INTEGRATION_FLOW_CONFIG_PARAMETER_VALUE'
```
