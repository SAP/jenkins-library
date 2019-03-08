# checkChangeInDevelopment

## Description

Checks if a Change Document in SAP Solution Manager is in status 'in development'. The change document id is retrieved from the git commit history. The change document id
can also be provided via parameter `changeDocumentId`. Any value provided as parameter has a higher precedence than a value from the commit history.

By default the git commit messages between `origin/master` and `HEAD` are scanned for a line like `ChangeDocument : <changeDocumentId>`. The commit
range and the pattern can be configured. For details see 'parameters' table.

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `changeDocumentId` | yes |  |  |
| `changeManagement/changeDocumentLabel` | no | `ChangeDocument\s?:` | regex pattern |
| `changeManagement/clientOpts` | no |  |  |
| `changeManagement/credentialsId` | no | `CM` |  |
| `changeManagement/endpoint` | yes |  |  |
| `changeManagement/git/format` | no | `%b` | see `git log --help` |
| `changeManagement/git/from` | no | `origin/master` |  |
| `changeManagement/git/to` | no | `HEAD` |  |
| `failIfStatusIsNotInDevelopment` | no | `true` | `true`, `false` |
| `script` | yes |  |  |

* `changeDocumentId` - The id of the change document to transport. If not provided, it is retrieved from the git commit history.
* `changeManagement/changeDocumentLabel` - A pattern used for identifying lines holding the change document id.
* `changeManagement/clientOpts` - Additional options for cm command line client, e.g. like JAVA_OPTS.
* `changeManagement/credentialsId` - The id of the credentials to connect to the Solution Manager. The credentials needs to be maintained on Jenkins.
* `changeManagement/endpoint` - The address of the Solution Manager.
* `changeManagement/git/format` - Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
* `changeManagement/git/from` - The starting point for retrieving the change document id
* `changeManagement/git/to` - The end point for retrieving the change document id
* `failIfStatusIsNotInDevelopment` - When set to `false` the step will not fail in case the step is not in status 'in development'.
* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in `script: this`. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections of the config.yml the configuration is possible:

| parameter | general | step | stage |
|-----------|---------|------|-------|
| `changeDocumentId` |  |  | X |
| `changeManagement/changeDocumentLabel` | X | X | X |
| `changeManagement/clientOpts` | X | X | X |
| `changeManagement/credentialsId` | X | X | X |
| `changeManagement/endpoint` | X | X | X |
| `changeManagement/git/format` | X | X | X |
| `changeManagement/git/from` | X | X | X |
| `changeManagement/git/to` | X | X | X |
| `failIfStatusIsNotInDevelopment` |  | X | X |
| `script` |  |  |  |

## Exceptions

* `AbortException`:
  * If the change id is not provided via parameter and if the change document id cannot be retrieved from the commit history.
  * If the change is not in status `in development`. In this case no exception will be thrown when `failIfStatusIsNotInDevelopment` is set to `false`.
* `IllegalArgumentException`:
  * If a mandatory property is not provided.

## Examples

The step is configured using a customer configuration file provided as
resource in an custom shared library.

```groovy
@Library('piper-lib-os@master') _

// the shared lib containing the additional configuration
// needs to be configured in Jenkins
@Library('foo@master') __

// inside the shared lib denoted by 'foo' the additional configuration file
// needs to be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this,
                             customDefaults: 'myConfig.yml'
```

Example content of `'resources/myConfig.yml'` in branch `'master'` of the repository denoted by
`'foo'`:

```yaml
general:
  changeManagement:
    changeDocumentLabel: 'ChangeDocument\s?:'
    cmClientOpts: '-Djavax.net.ssl.trustStore=<path to truststore>'
    credentialsId: 'CM'
    endpoint: 'https://example.org/cm'
    git:
      from: 'HEAD~1'
      to: 'HEAD'
      format: '%b'
```

The properties configured in section `'general/changeManagement'` are shared between all change managment related steps.

The properties can also be configured on a per-step basis:

```yaml
  [...]
  steps:
    checkChangeInDevelopment:
      changeManagement:
        endpoint: 'https://example.org/cm'
        [...]
      failIfStatusIsNotInDevelopment: true
```

The parameters can also be provided when the step is invoked:

```groovy
    // simple case. All mandatory parameters provided via
    // configuration, changeDocumentId provided via commit
    // history
    checkChangeInDevelopment script:this
```

```groovy
    // explict endpoint provided, we search for changeDocumentId
    // starting at the previous commit (HEAD~1) rather than on
    // 'origin/master' (the default).
    checkChangeInDevelopment script:this
                             changeManagement: [
                               endpoint: 'https:example.org/cm'
                               git: [
                                 from: 'HEAD~1'
                               ]
                             ]
```
