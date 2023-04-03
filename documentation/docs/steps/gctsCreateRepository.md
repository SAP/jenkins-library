# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step creates a local Git repository on an ABAP server.
More information about the [Git-enabled Change and Transport System (gCTS)](https://help.sap.com/docs/ABAP_PLATFORM_NEW/4a368c163b08418890a406d413933ba7/f319b168e87e42149e25e13c08d002b9.html).

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
