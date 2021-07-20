# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactTriggerIntegrationTest script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactTriggerIntegrationTest:
    iFlowServiceKeyCredentialsId: 'MY_IFLOW_SERVICE_KEY'
    integrationFlowId: 'INTEGRATION_FLOW_ID'
    contentType: 'text/plain'
    messageBodyPath: 'myIntegrationsTest/testBody'
```
