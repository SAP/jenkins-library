# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactUnDeploy script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactUnDeploy:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'MY_INTEGRATION_FLOW_ID'
```
