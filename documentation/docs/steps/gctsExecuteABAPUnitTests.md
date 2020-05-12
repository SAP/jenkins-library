# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step will execute every unit test associated with a package belonging to the specified local repository on an ABAP system.
Learn more about the SAP git-enabled Central Transport Sytem (gCTS) [here](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/201909.001/en-US/f319b168e87e42149e25e13c08d002b9.html). With gCTS, ABAP developments on ABAP servers can be maintained in Git repositories.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a Jenkinsfile.

```groovy
gctsExecuteABAPUnitTests(
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
  gctsExecuteABAPUnitTests:
    host: "https://abap.server.com:port"
    client: "000"
    username: "ABAPUsername"
    password: "ABAPPassword"
    repository: "myrepo"
```
