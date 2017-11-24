# centralPipelineLoad

## Description
Loads a pipeline from a git repository. The idea is to set up a pipeline job in Jenkins that loads a minimal pipeline, which in turn loads the shared library and then uses this step to load the actual pipeline.

## Prerequisites

none

## Parameters

| parameter          | mandatory | default         | possible values |
| -------------------|-----------|-----------------|-----------------|
| `repoUrl`          | yes       |                 |                 |
| `branch`           | no        | 'master'        |                 |
| `jenkinsfilePath`  | no        | 'Jenkinsfile'   |                 |
| `credentialsId`    | no        | An empty String |                 |

* `repoUrl` The url to the git repository of the pipeline to be loaded.
* `branch` The branch of the git repository from which the pipeline should be checked out.
* `jenkinsfilePath` The path to the Jenkinsfile, inside the repository, to be loaded.
* `credentialsId` The Jenkins credentials containing user and password needed to access a private git repository.

## Return value

none

## Side effects

The Jenkinsfile is checked out to a temporary folder in the Jenkins workspace. This folder starts with 'pipeline-' followed by a random UUID.

## Exceptions

* `Exception`
    * If `repoUrl` is not provided.

## Example

```groovy
centralPipelineLoad repoUrl: "https://github.com/MyOrg/MyPipelineRepo.git", branch: 'feature1', jenkinsfilePath: 'path/to/Jenkinsfile', credentialsId: 'MY_REPO_CREDENTIALS'
```
