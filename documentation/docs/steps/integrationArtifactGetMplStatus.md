# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationArtifactGetMplStatus script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactGetMplStatus:
    cpiAPIServiceKeyCredentialId: 'MY_API_SERVICE_KEY'
    integrationFlowId: 'INTEGRATION_FLOW_ID'
    platform: cf
```
