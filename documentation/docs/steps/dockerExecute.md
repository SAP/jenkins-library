# dockerExecute

## Description

Content here is generated from corresponnding step, see `vars`.

## Parameters

Content here is generated from corresponnding step, see `vars`.

## Kubernetes support

If the Jenkins is setup on a Kubernetes cluster, then you can execute the closure inside a container of a pod by setting an environment variable `ON_K8S` to `true`. However, it will ignore `containerPortMappings`, `dockerOptions` and `dockerVolumeBind` values.

## Step configuration

Content here is generated from corresponnding step, see `vars`.

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
