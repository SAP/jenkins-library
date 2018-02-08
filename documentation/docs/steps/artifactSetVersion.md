# artifactSetVersion

## Description
The continuous delivery process requires that each build is done with a unique version number.

The version generated using this step will contain:

* Version (major.minor.patch) from descriptor file in master repository is preserved. Developers should be able to autonomously decide on increasing either part of this version number.
* Timestamp
* CommitId (by default the long version of the hash)

After conducting automatic versioning the new version is pushed as a new tag into the source code repository (e.g. GitHub)

## Prerequsites
none

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | no | empty `commonPipelineEnvironment` |  |
| buildTool | no | maven | maven, docker |
| dockerVersionSource | no  | `''`  | FROM, (ENV name),appVersion  |
| filePath | no | buildTool=`maven`: pom.xml <br />docker: Dockerfile |   |
| gitCommitId |  no | `GitUtils.getGitCommitId()`   |   |
| gitCredentialsId |  yes | as defined in custom configuration  |   |
| gitUserEMail | no |  |   |
| gitUserName | no |   |   |
| gitSshUrl | yes  |  |   |
| tagPrefix | no  | 'build_'  |   |
| timestamp | no  |  current time in format according to `timestampTemplate`  |   |
| timestampTemplate | no | `%Y%m%d%H%M%S` |   |
| versioningTemplate | no | depending on `buildTool`<br />maven: `${version}-${timestamp}${commitId?"_"+commitId:""}`  |   |

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `buildTool` defines the tool which is used for building the artifact.
* `dockerVersionSource` specifies the source to be used for the main version which is used for generating the automatic version.

    * This can either be the version of the base image - as retrieved from the `FROM` statement within the Dockerfile, e.g. `FROM jenkins:2.46.2`
    * Alternatively the name of an environment variable defined in the Docker image can be used which contains the version number, e.g. `ENV MY_VERSION 1.2.3`
    * The third option `appVersion` applies only to the artifactType `appContainer`. Here the version of the app which is packaged into the container will be used as version for the container itself.

* Using `filePath` you could define a custom path to the descriptor file.
* `gitCommitId` defines the version prefix of the automatically generated version. By default it will take the long commitId hash. You could pass any other string (e.g. the short commitId hash) to be used. In case you don't want to have the gitCommitId added to the automatic versioning string you could set the value to an empty string: `''`.
* `gitCredentialsId`defines the ssh git credentials to be used for writing the tag.
* The parameters `gitUserName` and `gitUserEMail` allow to overwrite the global git settings available on your Jenkins server
* `gitSshUrl` defines the git ssh url to the source code repository.
* `tagPrefix` defines the prefix wich is used for the git tag which is written during the versioning run.
* `timestamp` defines the timestamp to be used in the automatic version string. You could overwrite the default behavior by explicitly setting this string.

## Step configuration
Following parameters can also be specified as step parameters using the global configuration file:

* `artifactType`
* `buildTool`
* `dockerVersionSource`
* `filePath`
* `gitCredentialsId`
* `gitUserEMail`
* `gitUserName`
* `gitSshUrl`
* `tagPrefix`
* `timestamp`
* `timestampTemplate`
* `versioningTemplate`

## Explanation of pipeline step

Pipeline step:

```groovy
artifactSetVersion script: this, buildTool: 'maven'
```


