# GitUtils

## Description
Provides git utility functions.

## Constructors

### GitUtils()

Default no-argument constructor. Instances of the GitUtils class does not hold any instance specific state.

#### Example

```groovy
new GitUtils()
```

## Method Details

### retrieveGitCoordinates(script)

#### Description
Retrieves the git-remote-url and git-branch. The parameters 'GIT_URL' and 'GIT_BRANCH' are retrieved from Jenkins job configuration. If these are not set, the git-url and git-branch are retrieved from the same repository where the Jenkinsfile resides.

#### Parameters

* `script` The script calling the method. Basically the `Jenkinsfile`. It is assumed that the script provides access to the parameters defined when launching the build, especially `GIT_URL` and `GIT_BRANCH`.

#### Return value

A map containing git-url and git-branch: `[url: gitUrl, branch: gitBranch]`

## Exceptions

* `AbortException`
    * If there is no SCM present. This happens when the there is no `Jenkinsfile`, when the pipeline is defined in the job configuration.
    * If only one of `GIT_URL`,  `GIT_BRANCH` is set in the Jenkins job configuration.

#### Example

```groovy
def gitCoordinates = new GitUtils().retrieveGitCoordinates(this)
def gitUrl = gitCoordinates.url
def gitBranch = gitCoordinates.branch
```
