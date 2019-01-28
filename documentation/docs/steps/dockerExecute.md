# dockerExecute

## Description

Executes a closure inside a docker container with the specified docker image.
The workspace is mounted into the docker image.
Proxy environment variables defined on the Jenkins machine are also available in the Docker container.

## Parameters
| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|containerCommand|no|||
|containerPortMappings|no|||
|containerShell|no|||
|dockerEnvVars|no|`[:]`||
|dockerImage|no|`''`||
|dockerName|no|||
|dockerOptions|no|`''`||
|dockerAlwaysPullImage|no|true|boolean value: `true`, `false` |
|dockerVolumeBind|no|`[:]`||
|dockerWorkspace|no|||
|jenkinsKubernetes|no|`[jnlpAgent:s4sdk/jenkins-agent-k8s:latest]`||
|sidecarEnvVars|no|||
|sidecarImage|no|||
|sidecarName|no|||
|sidecarOptions|no|||
|sidecarAlwaysPullImage|no|true|boolean value: `true`, `false` |
|sidecarVolumeBind|no|||
|sidecarWorkspace|no|||

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `containerCommand`: only used in case exeuction environment is Kubernetes, allows to specify start command for container created with dockerImage parameter to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
* `containerPortMappings`: Map which defines per docker image the port mappings, like `containerPortMappings: ['selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]]`
* `containerShell`: only used in case exeuction environment is Kubernetes, allows to specify the shell to be used for execution of commands
* `dockerEnvVars`: Environment variables to set in the container, e.g. [http_proxy:'proxy:8080']
* `dockerImage`: Name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
* `dockerName`: Kubernetes case: Name of the container launching `dockerImage`, SideCar: Name of the container in local network
* `dockerOptions` Docker options to be set when starting the container. It can be a list or a string.
* `dockerAlwaysPullImage`: Set this to 'false' to bypass a docker image pull. Usefull during development process. Allows testing of images which are available in the local registry only.
* `dockerVolumeBind` Volumes that should be mounted into the container.
* `dockerWorkspace`: only relevant for Kubernetes case: specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`
* `sidecarEnvVars` defines environment variables for the sidecar container, similar to `dockerEnvVars`
* `sidecarImage`: Name of the docker image of the sidecar container. Do not provide this value if no sidecar container is required.
* `sidecarName`: as `dockerName` for the sidecar container
* `sidecarOptions`: as `dockerOptions` for the sidecar container
* `sidecarAlwaysPullImage`: Set this to 'false' to bypass a docker image pull. Usefull during development process. Allows testing of images which are available in the local registry only.
* `sidecarVolumeBind`: as `dockerVolumeBind` for the sidecar container
* `sidecarWorkspace`: as `dockerWorkspace` for the sidecar container

## Kubernetes support

If the Jenkins is setup on a Kubernetes cluster, then you can execute the closure inside a container of a pod by setting an environment variable `ON_K8S` to `true`. However, it will ignore `containerPortMappings`, `dockerOptions` and `dockerVolumeBind` values.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|containerPortMappings||X|X|
|dockerEnvVars||X|X|
|dockerImage||X|X|
|dockerName||X|X|
|dockerOptions||X|X|
|dockerAlwaysPullImage||X|X|
|dockerVolumeBind||X|X|
|dockerWorkspace||X|X|
|jenkinsKubernetes|X|||
|sidecarEnvVars||X|X|
|sidecarImage||X|X|
|sidecarName||X|X|
|sidecarOptions||X|X|
|sidecarAlwaysPullImage||X|X|
|sidecarVolumeBind||X|X|
|sidecarWorkspace||X|X|

## Side effects

none

## Exceptions

none

## Example 1: Run closure inside a docker container

```groovy
dockerExecute(dockerImage: 'maven:3.5-jdk-7'){
    sh "mvn clean install"
}
```

## Example 2: Run closure inside a container in a kubernetes pod

```sh
# set environment variable
export ON_K8S=true"
```

```groovy
dockerExecute(script: this, dockerImage: 'maven:3.5-jdk-7'){
    sh "mvn clean install"
}
```

In the above example, the `dockerEcecute` step will internally invoke [dockerExecuteOnKubernetes](dockerExecuteOnKubernetes.md) step and execute the closure inside a pod.

## Example 3: Run closure inside a container which is attached to a sidecar container (as for example used in [seleniumExecuteTests](seleniumExecuteTests.md)

```groovy
dockerExecute(
        script: script,
        containerPortMappings: [containerPortMappings:'selenium/standalone-chrome':[containerPort: 4444, hostPort: 4444]],
        dockerImage: 'node:8-stretch',
        dockerName: 'node',
        dockerWorkspace: '/home/node',
        sidecarImage: 'selenium/standalone-chrome',
        sidecarName: 'selenium',
) {
    git url: 'https://github.wdf.sap.corp/XXXXX/WebDriverIOTest.git'
    sh '''npm install
          node index.js
    '''
}
```
