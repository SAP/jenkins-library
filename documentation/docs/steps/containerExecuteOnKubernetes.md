# containerExecuteInsidePod

## Description

Executes a closure inside a pod with multiple containers. The containers can be chosen as part of the closure.

## Prerequisites 
* The Jenkins should be running on kubernetes.
* An environment variable `ON_K8S` should be created on Jenkins and initialized to `true`.
* The step should be invoked only inside a [`podTemplate`](https://jenkins.io/doc/pipeline/steps/kubernetes/#podtemplate-define-a-podtemplate-to-use-in-the-kubernetes-plugin)
 
## Parameters

| parameter          | mandatory | default                           | possible values            |
| -------------------|-----------|-----------------------------------|----------------------------|
| `script`      | no        | empty `globalPipelineEnvironment`                                |                            |
| `containerMap`      | yes        |                                | `['maven:3.5-jdk-7':'maven']`                           |
| `dockerEnvVars`    | no        | [:]                               |                            |
| `dockerWorkspace`    | no        | ''                                |                            |

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for storing the measured duration.
* `containerMap` A map containing docker image as key and respective container name as value. 
* `dockerEnvVars` Environment variables to set in the container, e.g. [http_proxy:'proxy:8080']
* `dockerWorkspace` Docker options to be set when starting the container. It can be a list or a string.

## General configuration
This step adds a `jnlpAgent` container to the list of containers in a pod and uses 's4sdk/jenkins-agent-k8s:latest' docker image. To use the custom `jnlpAgent` agent, it can be configured in the general configuration section of the pipeline configuration.

```yaml
general:
  jenkinsKubernetes:
    jnlpAgent: 's4sdk/jenkins-agent-k8s:latest'
```

## Exceptions

None

## Example

```groovy
containerExecuteInsidePod(
                         script: this, 
                         containerMap: ['s4sdk/docker-node-browsers':'node','maven:3.5-jdk-8-alpine':'maven'], 
                         dockerWorkspace: '/var/build') 
                         {
                            container(name: 'maven') {
                                    sh "mvn clean install"
                                }
                            container(name: 'node') {
                                  sh "npm install"
                               }
                        }
```




