# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## Kubernetes support

If the Jenkins is setup on a Kubernetes cluster, then you can execute the closure inside a container of a pod by setting an environment variable `ON_K8S` to `true`. However, it will ignore `containerPortMappings`, `dockerOptions` and `dockerVolumeBind` values.
 `dockerExecute` step will internally invoke [dockerExecuteOnKubernetes](dockerExecuteOnKubernetes.md) step and execute the closure inside a pod.

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Exceptions

none

## Pulling images in an non-anonymous way

!!! warning "Credentials are stored by default unencrypted on disk"
    When accessing a docker registry with credentials for pulling
    images your credentials for access the docker registry
    are stored in plain text on disk for a short amount of time.
    There will be a corresponding log message with level "warning" in
    the job log.
    In order to avoid having the credentials written to disk, you
    should configure a password helper. The log message mentioned
    previously contains a link to a page explaining how a password helper
    can be configured.
    Having the credentials written to disk is not recommended.
    In addition, we don't recommend using personalised accounts for CI but rather dedicated "technical" users.

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

In the above example, the `dockerExecute` step will internally invoke [dockerExecuteOnKubernetes](dockerExecuteOnKubernetes.md) step and execute the closure inside a pod.

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
    git url: 'https://github.com/XXXXX/WebDriverIOTest.git'
    sh '''npm install
          node index.js
    '''
}
```
