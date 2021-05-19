# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## Administering the Transport Request ID by Git Commit Messages

The transport request ID identifies a container in the ABAP development system that can be used to document and transport changes within the landscape.
If artifacts created by a pipeline are to be loaded into such a container, then the transport request ID must be provided in an upload step. See [transportRequestUploadSOLMAN](transportRequestUploadSOLMAN.md).

`transportRequestReqIDFromGit` allows to retrieve the ID from a commit message of the Git repository of the project. This allows the developer to address the transport request without having to change the setup of the pipeline.
The developer only has to make sure that the ID is unique in the defined search range.

### Specifying the Git Commit Message

The Git commit messages (`git log`) of the project are searched for lines that follow a defined pattern.
The pattern is specified by the label _transportRequestLabel_ (default=`TransportRequest`).
Behind the label a colon, any blanks, and the identifier are expected.

```
Upload - define the transport request ID

    TransportRequest: ABCD10005E
```

The Git commit messages to be considered are determined by the parameters _gitFrom_ (default=`origin/master`) and _gitTo_ (default=`HEAD`).
The naming follows the Git revision range representation `git log <gitFrom>..<gitTo>`.
All commit messages accessible from _gitTo_ but not from _gitFrom_ are taken into account.
If the scanner detects multiple IDs, it fails. So the commit range has to be chosen accordingly.

In case of a pull request of a feature branch, the default should be sufficient as long as the transport request isn't changed.
Only the commits (`HEAD`) that have not yet entered the main branch `origin/master` would be scanned.

If uploading from the main branch, it must be assumed that former transport request IDs may already contained in the history. In this case the new ID should be maintained in the `HEAD` and
_gitFrom_ be set to `HEAD~1`.

```yaml
steps:
  transportRequestReqIDFromGit:
    gitFrom: 'HEAD~1'
```

### Executed on Jenkins Master

This step is executed on the Jenkins master only. On the master the project's Git repository is fully available. If kubernetes is used, the Git repository would have to be stashed. Depending on the size of the repository, this would be quite expensive or not possible at all.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
transportRequestReqIDFromGit( script: this )
```
