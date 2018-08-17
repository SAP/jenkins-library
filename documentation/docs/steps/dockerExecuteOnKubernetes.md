# dockerExecuteOnKubernetes

## Description

Executes a closure inside a container in a kubernetes pod. Proxy environment variables defined on the Jenkins machine are also available in the container.

## Prerequisites 
* The Jenkins should be running on kubernetes.
* An environment variable `ON_K8S` should be created on Jenkins and initialized to `true`.
 
## Parameters

| parameter          | mandatory | default                           | possible values            |
| -------------------|-----------|-----------------------------------|----------------------------|
| `script`           | no        | empty `globalPipelineEnvironment` |                            |
| `dockerImage`      | yes       |                                   |                            |
| `dockerEnvVars`    | no        | [:]                               |                            |
| `dockerWorkspace`  | no        | ''                                |                            |
| `containerMap`     | no        | [:]                               |                            |         
* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `dockerImage` Name of the docker image that should be used. If empty, Docker is not used.
* `dockerEnvVars` Environment variables to set in the container, e.g. [http_proxy:'proxy:8080']
* `dockerWorkspace` Docker options to be set when starting the container. It can be a list or a string.
* `containerMap` A map of docker image to the name of the container. The pod will be created with all the images from this map and they are labled based on the value field of each map entry.
   Ex `['maven:3.5-jdk-8-alpine': 'mavenExecute', 'checkJMeter': 'checkJMeter', 'deployToCfWithCli': 'deployToCfWithCli']`

## Step configuration
None

## Exceptions

None

## Example

```groovy
withEnv(["ON_K8S=true"]){
dockerExecuteOnKubernetes(script: script,
                          dockerImage: 'maven:3.5-jdk-7')
                          {
                            sh "mvn clean install" 
                          }
                        }
```




