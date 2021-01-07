# ${docGenStepName}

## ${docGenDescription}

With this step you can deploy a integration flow artifact in to SAP Cloud Platform integration runtime using OData API.

Learn more about the SAP Cloud Integration remote API for deploying an integration artifact [here](https://help.sap.com/viewer/368c481cd6954bdfa5d0435479fd4eaf/Cloud/en-US/08632076a1114bc1b6a1ecafef8f0178.html).

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
deployIntegrationArtifact script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  deployIntegrationArtifact:
    cpiCredentialsId: 'MY_CPI_OAUTH_CREDENTIALSID_IN_JENKINS'
    integrationFlowId: 'MY_INTEGRATION_FLOW_NAME'
    integrationFlowVersion: 'MY_INTEGRATION_FLOW_VERSION'
    platform: cf
    host: https://CPI_HOST_ITSPACES_URL
    oAuthTokenProviderUrl: https://CPI_HOST_OAUTH_URL
```
