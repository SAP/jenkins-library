# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactGetServiceEndpoint script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactGetServiceEndpoint:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'MY_INTEGRATION_FLOW_ID'
```
