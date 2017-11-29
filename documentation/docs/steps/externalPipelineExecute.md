# externalPipelineExecute

## Description
Loads a pipeline from a git repository. The idea is to set up a pipeline job in Jenkins that loads a minimal pipeline, which in turn loads the shared library and then uses this step to load the actual pipeline.

## Prerequisites

none

## Parameters

| parameter          | mandatory | default         | possible values |
| -------------------|-----------|-----------------|-----------------|
| `repoUrl`          | yes       |                 |                 |
| `branch`           | no        | 'master'        |                 |
| `path`             | no        | 'Jenkinsfile'   |                 |
| `credentialsId`    | no        | An empty String |                 |

* `repoUrl` The url to the git repository of the pipeline to be loaded.
* `branch` The branch of the git repository from which the pipeline should be checked out.
* `path` The path to the Jenkinsfile, inside the repository, to be loaded.
* `credentialsId` The Jenkins credentials containing user and password needed to access a private git repository.

## Return value

none

## Side effects

none

## Exceptions

* `Exception`
    * If `repoUrl` is not provided.

## Example

```groovy
externalPipelineExecute repoUrl: "https://github.com/MyOrg/MyPipelineRepo.git", branch: 'feature1', path: 'path/to/Jenkinsfile', credentialsId: 'MY_REPO_CREDENTIALS'
```
