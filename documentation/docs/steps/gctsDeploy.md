# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

With this step you can deploy a commit from a remote Git repository to a local repository on an ABAP server. If no `commit` parameter is specified, this step will pull the latest commit available on the remote repository.
Learn more about the SAP git-enabled Central Transport Sytem (gCTS) [here](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/201909.001/en-US/f319b168e87e42149e25e13c08d002b9.html). With gCTS, ABAP developments on ABAP servers can be maintained in Git repositories.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
gctsDeploy(
  script: this,
  host: "https://abap.server.com:port",
  client: "000",
  abapCredentialsId: 'ABAPUserPasswordCredentialsId',
  repository: "myrepo"
)
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  gctsDeploy:
    host: "https://abap.server.com:port"
    client: "000"
    username: "ABAPUsername"
    password: "ABAPPassword"
    repository: "myrepo"
```
