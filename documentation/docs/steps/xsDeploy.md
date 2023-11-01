# ${docGenStepName}

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
