# checkChangeInDevelopment

## Description

Checks if a Change Document in SAP Solution Manager is in status 'in development'. The change document id is retrieved from the git commit history. The change document id
can also be provided via parameter `changeDocumentId`. Any value provided as parameter has a higher precedence than a value from the commit history.

By default the git commit messages between `origin/master` and `HEAD` are scanned for a line like `ChangeDocument : <changeDocumentId>`. The commit
range and the pattern can be configured. For details see 'parameters' table.

In case the change is not in status 'in development' an `hudson.AbortException` is thrown. In case `failIfStatusIsNotInDevelopment`
is set to `false`, no `hudson.AbortException` will be thrown. In this case there is only a message in the log stating the change is not in status 'in development'.

## Prerequisites

* **[Change Management Client 2.0.0 or compatible version](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/)** - available for download on Maven Central.

## Parameters

Content here is generated from corresponnding step, see `vars`.

## Step configuration

Content here is generated from corresponnding step, see `vars`.

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
