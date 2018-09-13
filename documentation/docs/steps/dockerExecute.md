# dockerExecute

## Description

Executes a closure inside a docker container with the specified docker image. 
The workspace is mounted into the docker image.
Proxy environment variables defined on the Jenkins machine are also available in the Docker container.

## Parameters

| parameter          | mandatory | default                           | possible values            |
| -------------------|-----------|-----------------------------------|----------------------------|
| `script`           | no        | empty `globalPipelineEnvironment` |                            |
| `dockerImage`      | no        | ''                                |                            |
| `dockerEnvVars`    | no        | [:]                               |                            |
| `dockerOptions`    | no        | ''                                |                            |
| `dockerVolumeBind` | no        | [:]                               |                            |

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `dockerImage` Name of the docker image that should be used. If empty, Docker is not used.
* `dockerEnvVars` Environment variables to set in the container, e.g. [http_proxy:'proxy:8080']
* `dockerOptions` Docker options to be set when starting the container. It can be a list or a string.
* `dockerVolumeBind` Volumes that should be mounted into the container.


## Kubernetes support
If the Jenkins is setup on a Kubernetes cluster, then you can execute the closure inside a container of a pod by setting an environment variable `ON_K8S` to `true`. However, it will ignore both `dockeOptions` and `dockerVolumeBind` values.

## Step configuration
none

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




