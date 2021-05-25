# ${docGenStepName}

## ${docGenDescription}

## Administering the Change Document ID by Git Commit Messages

A `change document` documents activities in the change process.
To upload an artifact into a transport request, the Solution Manager expects the ID of an assigned change document. For more information, see [transportRequestUploadSOLMAN](transportRequestUploadSOLMAN.md).

`transportRequestDocIDFromGit` allows you to retrieve the commit message ID in the Git repository of your project. This way, you can address the change document without having to change the setup of your pipeline.
Please make sure that the ID is unique in the defined search range.

### Specifying the Git Commit Message

`transportRequestDocIDFromGit` searches for lines that follow a defined pattern in the Git commit messages (`git log`) of your project.
If necessary, specify the pattern with the label _changeDocumentLabel_ (default=`ChangeDocument`).
Behind the label, enter a colon, blank spaces, and the identifier.

```
Upload - define the change document ID

    ChangeDocument: 1000001234
```

### Specifying the Git Commit Range

The Git commit messages to be considered are determined by the parameters _gitFrom_ (default=`origin/master`) and _gitTo_ (default=`HEAD`).
The naming follows the Git revision range notation `git log <gitFrom>..<gitTo>`.
All commit messages accessible from _gitTo_ but not from _gitFrom_ are taken into account.
If the scanner detects multiple IDs, it fails. So the commit range has to be chosen accordingly.

In case you want to retrieve the ID within the scope of a pull request, the default values `HEAD` and `origin/master` should be sufficient provided that

* you committed the change document ID into the pull request
* you did not merge the `origin/master` after that

```
o 3d97415 (origin/master) merged change
|
| o d99fbf7 (HEAD) feature fixes
| |
| o 5c380ea ChangeDocument: 1000001234
| |
| o 0e82d9b new feature
|/
o 4378bb4 merged change 

```

In case you want to retrieve the ID from the main branch, former change document IDs may already be in the history.
Adjust _gitFrom_ so that it points to a commit before your ID definition.

```yaml
steps:
  transportRequestDocIDFromGit:
    gitFrom: '4378bb4'
```

```
o d99fbf7 (origin/master) feature fixes
|
o 5c380ea adjust config.yaml
|           ChangeDocument: 1000001234
|
o 0e82d9b new feature
|
o 4378bb4 merged change 

```

### Executed on Jenkins Master

This step is executed on the Jenkins master only. On the master the project's Git repository is fully available. If kubernetes is used, the Git repository would have to be stashed. Depending on the size of the repository, this would be quite expensive or not possible at all.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
transportRequestDocIDFromGit( script: this )
```
