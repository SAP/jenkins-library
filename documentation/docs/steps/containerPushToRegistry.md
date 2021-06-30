# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

You need to have a valid user with write permissions in the target docker registry.

Credentials for the target docker registry have been configured in Jenkins with a dedicated Id.

You can create the credentials in your Jenkins<br />
via _Jenkins_ -> _Credentials_ -> _System_ -> _Global credentials (unrestricted)_ -> _Add Credentials_ ->

* Kind: _Username with Password_
* ID: specify id which you then use for the configuration of `dockerCredentialsId` (see below)

## Example

Usage of pipeline step:

**OPTION A:** To pull a Docker image from an existing docker registry and push to a different docker registry:

```groovy
containerPushToRegistry script: this,
                        dockerCredentialsId: 'myTargetRegistryCredentials',
                        sourceRegistryUrl: 'https://mysourceRegistry.url',
                        sourceImage: 'path/to/mySourceImageWith:tag',
                        dockerRegistryUrl: 'https://my.target.docker.registry:50000'
```

**OPTION B:** To push a locally built docker image into the target registry (only possible when a Docker daemon is available on your Jenkins node):

```groovy
containerPushToRegistry script: this,
                        dockerCredentialsId: 'myTargetRegistryCredentials',
                        dockerImage: 'path/to/myImageWith:tag',
                        dockerRegistryUrl: 'https://my.target.docker.registry:50000'
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}
