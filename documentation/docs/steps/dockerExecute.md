# dockerExecute

## Description

Executes a closure inside a docker container with the specified docker image. 
The workspace is mounted into the docker image.
Proxy environment variables defined on the Jenkins machine are also available in the Docker container.

## Parameters

| parameter          | mandatory | default                           | possible values            |
| -------------------|-----------|-----------------------------------|----------------------------|
| `dockerImage`      | no        | ''                                |                            |
| `dockerEnvVars`    | no        | [:]                               |                            |
| `dockerOptions`    | no        | ''                                |                            |
| `dockerVolumeBind` | no        | [:]                               |                            |

* `dockerImage` Name of the docker image that should be used. If empty, Docker is not used.
* `dockerEnvVars` Environment variables to set in the container, e.g. [http_proxy:'proxy:8080']
* `dockerOptions` Docker options to be set when starting the container. It can be a list or a string.
* `dockerVolumeBind` Volumes that should be mounted into the container.

## Step configuration
None

## Exceptions

None

## Example

```groovy
dockerExecute(dockerImage: 'maven:3.5-jdk-7'){
    sh "mvn clean install"
}
```




