# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

With this step you can create a local git-enabled CTS (gCTS) repository on an ABAP server.
Learn more about the SAP Git-enabled Change & Transport Sytem (gCTS) [here](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/201909.001/en-US/f319b168e87e42149e25e13c08d002b9.html). With gCTS, ABAP developments on ABAP servers can be maintained in Git repositories.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
gctsCreateRepository(
  script: this,
  host: 'https://abap.server.com:port',
  client: '000',
  abapCredentialsId: 'ABAPUserPasswordCredentialsId',
  repository: 'myrepo',
  remoteRepositoryURL: 'https://github.com/user/myrepo',
  role: 'SOURCE',
  vSID: 'ABC'
  )
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  gctsCreateRepository:
    host: 'https://abap.server.com:port'
    client: '000'
    abapCredentialsId: 'ABAPUserPasswordCredentialsId'
    repository: 'myrepo'
    remoteRepositoryURL: 'https://github.com/user/myrepo'
    role: 'SOURCE'
    vSID: 'ABC'
```
