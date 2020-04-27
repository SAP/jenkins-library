# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

With this step you can clone a remote Git repository to a local repository on an ABAP server. To be able to execute this step, the corresponding local repository has to exist on the local ABAP system.
Learn more about gCTS [here](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/201909.001/en-US/f319b168e87e42149e25e13c08d002b9.html).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
gctsCloneRepository(
  script: this,
  host: "https://abap.server.com:port",
  client: "000",
  abapCredentialsId: 'ABAPUserPasswordCredentialsId',
  repository: "myrepo"
  )
```

Example configuration for the use in a yaml config file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  gctsCloneRepository:
    host: "https://abap.server.com:port"
    client: "000"
    username: "ABAPUsername"
    password: "ABAPPassword"
    repository: "myrepo"
```
