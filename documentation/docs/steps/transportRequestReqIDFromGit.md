# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

!!! note "Executed on Jenkins Master"
    This step is executed on the Jenkins master only. On the master node the project's Git repository is fully available. If kubernetes is used, the Git repository would have to be stashed. Depending on the size of the repository, this would be quite expensive or not possible at all.

## Administering the Transport Request ID by Git Commit Messages

The `transport request ID` identifies a container in the ABAP development system that can be used to document and transport changes within the landscape.
If you upload your pipeline artifacts into such a container, please provide the transport request ID in an upload step.
See [transportRequestUploadSOLMAN](transportRequestUploadSOLMAN.md).

With `transportRequestReqIDFromGit`  you can retrieve the transport request ID from the commit message of the Git repository of your project. This way, you can address the transport request without having to change the setup of your pipeline.
Please make sure that the ID is unique in the defined search range.

## General Purpose Pipeline Init Stage

The step can also be configured via General Purpose Pipeline in Init stage using the config.yml as follows:

```yaml
stages:
  Init:
    transportRequestReqIDFromGit: true
```

This will initialize the step within the Init stage of the pipeline and retrieve the `transportRequestId` from the git commit history.

### Specifying the Git Commit Message

`transportRequestReqIDFromGit` searches for lines that follow a defined pattern in the Git commit messages (`git log`) of your project.
Only if necessary, specify the pattern with the label _transportRequestLabel_ (default=`TransportRequest`).
Behind the label, enter a colon, blank spaces, and the identifier.

```
Upload - define the transport request ID

    TransportRequest: ABCD10005E
```

### Specifying the Git Commit Range

The Git commit messages to be considered are determined by the parameters _gitFrom_ (default=`origin/master`) and _gitTo_ (default=`HEAD`).
The naming follows the Git revision range notation `git log <gitFrom>..<gitTo>`.
All commit messages accessible from _gitTo_ but not from _gitFrom_ are taken into account.
Choose the commit range accordingly, as the detection of multiple IDs causes the scan to fail.

Keep the default values `HEAD` and `origin/master` in case you want to retrieve the ID within the scope of a pull request.
The default values should be sufficient provided that

* you commit the transport request ID into the pull request
* you do not merge the `origin/master` before the scan
* you do not change the transport request ID while developing

This way, only the commits (`HEAD`) that have not yet entered the main branch `origin/master` are scanned.

```
o 3d97415 (origin/master) merged last change
|
| o d99fbf7 (HEAD) feature fixes
| |
| o 5c380ea TransportRequest: ABCD10001E
| |
| o 0e82d9b new feature
|/
o 4378bb4 last change
```

If you want to retrieve the ID from the main branch, be aware that former transport request IDs may already be in the history.
Adjust _gitFrom_ so that it points to a commit before your ID definition.

```yaml
steps:
  transportRequestReqIDFromGit:
    gitFrom: '4378bb4'
```

```
o 3d97415 (origin/master) merge new feature
|
o d99fbf7 feature fixes
|
o 5c380ea adjust config.yaml
|           TransportRequest: ABCD10001E
|
o 0e82d9b new feature
|
o 4378bb4 merged last change
```

Define _gitTo_, if it cannot be ruled out that further transport request IDs have been merged in parallel.

```yaml
steps:
  transportRequestReqIDFromGit:
    gitFrom: '4378bb4'
    gitTo: 'd99fbf7'
```

```
o 3d97415 (origin/master) merge new feature
|\
. o d99fbf7 feature fixes
. |
. o 5c380ea adjust config.yaml
. |           TransportRequest: ABCD10001E
. |
. o 0e82d9b new feature
|/
o 4378bb4 merged last change
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
transportRequestReqIDFromGit( script: this )
```
