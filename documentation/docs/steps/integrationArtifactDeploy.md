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
integrationArtifactDeploy script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactDeploy:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'MY_INTEGRATION_FLOW_NAME'
```
