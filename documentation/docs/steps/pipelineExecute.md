# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

none

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Exceptions

* `Exception`
  * If `repoUrl` is not provided.

## Example

```groovy
pipelineExecute repoUrl: "https://github.com/MyOrg/MyPipelineRepo.git", branch: 'feature1', path: 'path/to/Jenkinsfile', credentialsId: 'MY_REPO_CREDENTIALS'
```
