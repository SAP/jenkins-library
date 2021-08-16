# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

If you provide a `commit ID`, the step deploys the content of the specified commit to the ABAP system. If you provide a `branch`, the step deploys the content of the specified branch. If you set the `rollback` parameter to *true*, the step returns to a working state of the repository, if the deployment of the specified commit or branch fails.
Learn more about the SAP Git-enabled Change & Transport System (gCTS) [here](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/201909.001/en-US/f319b168e87e42149e25e13c08d002b9.html). With gCTS, ABAP developments on ABAP servers can be maintained in Git repositories.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
gctsDeploy(
  script: this,
  host: 'https://abap.server.com:port',
  client: '000',
  abapCredentialsId: 'ABAPUserPasswordCredentialsId',
  repository: 'myrepo',
  remoteRepositoryURL: "https://remote.repository.url.com",
  role: 'SOURCE',
  vSID: 'ABC',
  branch: 'feature1',
  commit: '95952ec',
  scope: 'LASTACTION',
  rollback: true,
  configuration: [VCS_AUTOMATIC_PULL: 'FALSE',VCS_AUTOMATIC_PUSH: 'FALSE', CLIENT_VCS_LOGLVL: 'debug']
)
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  gctsDeploy:
    host: 'https://abap.server.com:port'
    client: '000'
    abapCredentialsId: 'ABAPUserPasswordCredentialsId'
    repository: 'myrepo'
    remoteRepositoryURL: "https://remote.repository.url.com"
    role: 'SOURCE'
    vSID: 'ABC'
    branch: 'feature2'
    commit: '0c9d330'
    scope: 'CRNTCOMMIT'
    rollback: false
    configuration:
        VCS_AUTOMATIC_PULL: "FALSE"
        VCS_AUTOMATIC_PUSH: "FALSE"
        CLIENT_VCS_LOGLVL: "debug"
```
