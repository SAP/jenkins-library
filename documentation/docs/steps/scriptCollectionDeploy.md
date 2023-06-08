# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
scriptCollectionDeploy script: this
```

Example of a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationArtifactDeploy:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    scriptCollectionId: 'MY_SCRIPT_COLLECTION_NAME'
```
