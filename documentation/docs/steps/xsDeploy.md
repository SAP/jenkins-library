# ${docGenStepName}

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Example

```groovy
xsDeploy
    script: this,
    mtaPath: 'path/to/archiveFile.mtar',
    credentialsId: 'my-credentials-id',
    apiUrl: 'https://example.org/xs',
    space: 'mySpace',
    org: 'myOrg'
```

Example configuration:

```yaml
steps:
  <...>
  xsDeploy:
    mtaPath: path/to/archiveFile.mtar
    credentialsId: my-credentials-id
    apiUrl: https://example.org/xs
    space: mySpace
    org: myOrg
```
