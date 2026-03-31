# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

If you provide a `commit ID`, the step deploys the content of the specified commit to the ABAP system. If you provide a `branch`, the step deploys the content of the specified branch. If you set the `rollback` parameter to *true*, the step returns to a working state of the repository, if the deployment of the specified commit or branch fails.
More information about the [Git-enabled Change and Transport System (gCTS)](https://help.sap.com/docs/ABAP_PLATFORM_NEW/4a368c163b08418890a406d413933ba7/f319b168e87e42149e25e13c08d002b9.html).

## Repository ID Resolution

The `repository` parameter is now **optional**. When omitted, the step automatically resolves the repository ID by querying all repositories on the ABAP system and matching against the `remoteRepositoryURL`.

### How it works

1. If `repository` is explicitly provided, it's used directly (backward compatible, fastest)
2. If `repository` is omitted, the step:
   - Calls the gCTS API to list all repositories on the ABAP system
   - Searches for a repository with matching `remoteRepositoryURL`
   - Uses the found repository ID
   - Errors if no match or multiple matches are found

### Advantages

- **Simpler configuration** - only the repository URL is needed
- **Works with any naming scheme** - repository ID doesn't need to match URL pattern
- **1:1 mapping guaranteed** - matches on actual configured URL

### Recommendation

For production pipelines, you may still explicitly provide the `repository` parameter for:

- **Faster execution** - avoids the list API call
- **Explicit documentation** - clearly shows which repository is used
- **Avoiding ambiguity** - if multiple repositories could exist with the same URL

### Error Scenarios

- **No matching repository**: If no repository exists with the specified URL, the step will fail with a clear error message suggesting to create the repository first or provide the `repository` parameter explicitly.

- **Multiple matching repositories**: If multiple repositories have the same remote URL (unusual but possible), the step will fail and list all matching repository IDs, asking you to specify which one to use via the `repository` parameter.

### Examples

**Traditional approach with explicit repository** (recommended for production):

```groovy
gctsDeploy(
  script: this,
  repository: 'myrepo',
  remoteRepositoryURL: "https://github.com/org/myrepo.git",
  host: HOST,
  client: CLIENT,
  abapCredentialsId: CREDS
)
```

**New approach with auto-resolved repository** (useful for development/testing):

```groovy
gctsDeploy(
  script: this,
  remoteRepositoryURL: "https://github.com/org/myrepo.git",
  host: HOST,
  client: CLIENT,
  abapCredentialsId: CREDS
)
```


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
  configuration: [VCS_AUTOMATIC_PULL: 'FALSE',VCS_AUTOMATIC_PUSH: 'FALSE',CLIENT_VCS_LOGLVL: 'debug'],
  queryparameters: [saml2: 'disabled']
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
    queryparameters:
        saml2: "disabled"
```
