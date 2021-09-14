# ${docGenStepName}

## ${docGenDescription}

### Additional hints

To run the `cnbBuild` with a different builder, you can specify the `dockerImage` parameter.
Without specifying it, the step will run with the `paketobuildpacks/builder:full` builder.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example 1

```groovy
cnbBuild(
    script: script,
    dockerConfigJsonCredentialsId: 'DOCKER_REGISTRY_CREDS',
    containerImageName: 'images/example',
    containerImageTag: 'v0.0.1',
    containerImageRegistryUrl: 'gcr.io'
)
```

## Example 2: User provided builder

```groovy
cnbBuild(
    script: script,
    dockerConfigJsonCredentialsId: 'DOCKER_REGISTRY_CREDS',
    dockerImage: 'paketobuildpacks/builder:base',
    containerImageName: 'images/example',
    containerImageTag: 'v0.0.1',
    containerImageRegistryUrl: 'gcr.io'
)
```
