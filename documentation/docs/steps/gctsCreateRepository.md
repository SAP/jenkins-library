# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

With this step you can create a local git-enabled CTS (gCTS) repository on an ABAP server.
Learn more about gCTS [here](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/201909.001/en-US/f319b168e87e42149e25e13c08d002b9.html).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a Jenkinsfile.

```groovy
gctsCreateRepository(
  script: this,
  host: "abap.server.com:port",
  client: "000",
  credentialsId: 'ABAPUserPasswordCredentialsId',
  repository: "myrepo",
  remoteRepositoryURL: "https://github.com/user/myrepo",
  role: "SOURCE",
  vSID: "ABC"
  )
```

Example configuration for the use in a yaml config file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  gctsCreateRepository:
    host: "abap.server.com:port"
    client: "000"
    username: "ABAPUsername"
    password: "ABAPPassword"
    repository: "myrepo"
    remoteRepositoryURL: "https://github.com/user/myrepo",
    role: "SOURCE",
    vSID: "ABC"
```
