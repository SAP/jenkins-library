# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* A SAP Cloud Platform ABAP Environment system is available. 
* On this system, a Communication User, a Communication System and a Communication Arrangement is setup for the Communication Scenario SAP_COM_0510: "SAP Cloud Platform ABAP Environment - Software Component Test Integration (SAP_COM_0510)".
* It is recommended to use the Jenkins credentials configuration for user and password handling.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
withCredentials([usernamePassword(credentialsId: 'myCredentialsId', usernameVariable: 'USER', passwordVariable: 'PASSWORD')]) {
    pullGitRepositoryToSapCloudPlatformAbapEnvironment(
        host : 'https://host.com', 
        repositoryName : '/DMO/GIT_REPOSITORY',
        username : "$USER",
        password : "$PASSWORD",
        script : this
    ) 
}
```
