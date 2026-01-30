# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

This step performs a rollback of commit(s) in a local ABAP system repository. If a `commit` parameter is specified, it will be used as the target commit for the rollback. If no `commit` parameter is specified and the remote repository domain is 'github.com', the last commit with the status 'success' will be used for the rollback. Otherwise, `gctsRollback` will roll back to the previously active commit in the local repository.
More information about the [Git-enabled Change and Transport System (gCTS)](https://help.sap.com/docs/ABAP_PLATFORM_NEW/4a368c163b08418890a406d413933ba7/f319b168e87e42149e25e13c08d002b9.html).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a Jenkinsfile.

```groovy
gctsRollback(
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
  gctsRollback:
    host: "https://abap.server.com:port"
    client: "000"
    abapCredentialsId: 'ABAPUserPasswordCredentialsId'
    repository: "myrepo"
```
