# ${docGenStepName}

## ${docGenDescription}

see [Examples](#examples)

## Prerequisites

When pushing to a container registry, you need to maintain the respective credentials in your Jenkins credentials store:

`cnbBuild` expects a Docker `config.json` file containing the credential information for registries.
You can create it like explained in the [protocodeExecuteScan Prerequisites section](https://www.project-piper.io/steps/protecodeExecuteScan/#prerequisites).

Please copy this file and upload it to your Jenkins for example<br />
via _Jenkins_ -> _Credentials_ -> _System_ -> _Global credentials (unrestricted)_ -> _Add Credentials_ ->

* Kind: _Secret file_
* File: upload your `config.json` file
* ID: specify id which you then use for the configuration of `dockerConfigJsonCredentialsId` (see below)

## ${docGenParameters}

### Additional hints

To run the `cnbBuild` with a different builder, you can specify the `dockerImage` parameter.
Without specifying it, the step will run with the `paketobuildpacks/builder:base` builder.

#### Default Excludes

When building images, these files/folders are excluded from the build by default:

* Piper binary: `piper`
* Piper configuration folder: `.pipeline`
* Git folder: `.git`

This behavior can be overwritten by using the respective sections in [`project.toml`](https://buildpacks.io/docs/reference/config/project-descriptor/). Keep in mind that by doing so, no default excludes will be applied by the `cnbBuild` step at all.

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Example 1: simple usage

```groovy
cnbBuild(
    script: this,
    dockerConfigJsonCredentialsId: 'DOCKER_REGISTRY_CREDS',
    containerImageName: 'images/example',
    containerImageTag: 'v0.0.1',
    containerRegistryUrl: 'gcr.io'
)
```

### Example 2: User provided builder

```groovy
cnbBuild(
    script: this,
    dockerConfigJsonCredentialsId: 'DOCKER_REGISTRY_CREDS',
    dockerImage: 'paketobuildpacks/builder:base',
    containerImageName: 'images/example',
    containerImageTag: 'v0.0.1',
    containerRegistryUrl: 'gcr.io'
)
```

### Example 3: User provided buildpacks

```groovy
cnbBuild(
    script: this,
    dockerConfigJsonCredentialsId: 'DOCKER_REGISTRY_CREDS',
    containerImageName: 'images/example',
    containerImageTag: 'v0.0.1',
    containerRegistryUrl: 'gcr.io',
    buildpacks: ['docker.io/paketobuildpacks/nodejs', 'paketo-community/build-plan']
)
```

### Example 4: Build environment variables

```groovy
cnbBuild(
    script: this,
    dockerConfigJsonCredentialsId: 'DOCKER_REGISTRY_CREDS',
    containerImageName: 'images/example',
    containerImageTag: 'v0.0.1',
    containerRegistryUrl: 'gcr.io',
    buildEnvVars: [
        "FOO": "BAR"
    ]
)
```
