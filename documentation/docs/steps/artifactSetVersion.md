# artifactSetVersion

## Description
The continuous delivery process requires that each build is done with a unique version number.

The version generated using this step will contain:

* Version (major.minor.patch) from descriptor file in master repository is preserved. Developers should be able to autonomously decide on increasing either part of this version number.
* Timestamp
* CommitId (by default the long version of the hash)

Optionally, but enabled by default, the new version is pushed as a new tag into the source code repository (e.g. GitHub).
If this option is chosen, git credentials and the repository URL needs to be provided.
Since you might not want to configure the git credentials in Jenkins, committing and pushing can be disabled using the `commitVersion` parameter as described below.
If you require strict reproducibility of your builds, this should be used.

## Prerequsites
none

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | yes |  |  |
| artifactType | no |  | 'appContainer' |
| buildTool | no | maven | docker, dlang, golang, maven, mta, npm, pip, sbt |
| commitVersion | no | `true` | `true`, `false` |
| dockerVersionSource | no  | `''`  | FROM, (ENV name),appVersion  |
| filePath | no | buildTool=`docker`: Dockerfile <br />buildTool=`dlang`: dub.json <br />buildTool=`golang`: VERSION <br />buildTool=`maven`: pom.xml <br />buildTool=`mta`: mta.yaml <br />buildTool=`npm`: package.json <br />buildTool=`pip`: version.txt <br />buildTool=`sbt`: sbtDescriptor.json|  |
| gitCommitId |  no | `GitUtils.getGitCommitId()`   |   |
| gitSshCredentialsId |  If `commitVersion` is `true` | as defined in custom configuration  |   |
| gitUserEMail | no |  |   |
| gitUserName | no |   |   |
| gitSshUrl | If `commitVersion` is `true` |  |   |
| tagPrefix | no  | 'build_'  |   |
| timestamp | no  |  current time in format according to `timestampTemplate`  |   |
| timestampTemplate | no | `%Y%m%d%H%M%S` |   |
| versioningTemplate | no |buildTool=`docker`: `${version}-${timestamp}${commitId?"_"+commitId:""}`<br> />buildTool=`dlang`: `${version}-${timestamp}${commitId?"+"+commitId:""}`<br />buildTool=`golang`:`${version}-${timestamp}${commitId?"+"+commitId:""}`<br />buildTool=`maven`: `${version}-${timestamp}${commitId?"_"+commitId:""}`<br />buildTool=`mta`: `${version}-${timestamp}${commitId?"+"+commitId:""}`<br />buildTool=`npm`: `${version}-${timestamp}${commitId?"+"+commitId:""}`<br />buildTool=`pip`: `${version}.${timestamp}${commitId?"."+commitId:""}`<br />buildTool=`sbt`: `${version}-${timestamp}${commitId?"+"+commitId:""}`|  |

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `artifactType` defines the type of the artifact.
* `buildTool` defines the tool which is used for building the artifact.
* `commitVersion` controls if the changed version is committed and pushed to the git repository. If this is enabled (which is the default), you need to provide `gitCredentialsId` and `gitSshUrl`.
* `dockerVersionSource` specifies the source to be used for the main version which is used for generating the automatic version.

    * This can either be the version of the base image - as retrieved from the `FROM` statement within the Dockerfile, e.g. `FROM jenkins:2.46.2`
    * Alternatively the name of an environment variable defined in the Docker image can be used which contains the version number, e.g. `ENV MY_VERSION 1.2.3`
    * The third option `appVersion` applies only to the artifactType `appContainer`. Here the version of the app which is packaged into the container will be used as version for the container itself.

* Using `filePath` you could define a custom path to the descriptor file.
* `gitCommitId` defines the version prefix of the automatically generated version. By default it will take the long commitId hash. You could pass any other string (e.g. the short commitId hash) to be used. In case you don't want to have the gitCommitId added to the automatic versioning string you could set the value to an empty string: `''`.
* `gitSshCredentialsId`defines the ssh git credentials to be used for writing the tag.
* The parameters `gitUserName` and `gitUserEMail` allow to overwrite the global git settings available on your Jenkins server
* `gitSshUrl` defines the git ssh url to the source code repository.
* `tagPrefix` defines the prefix wich is used for the git tag which is written during the versioning run.
* `timestamp` defines the timestamp to be used in the automatic version string. You could overwrite the default behavior by explicitly setting this string.

## Step configuration
The following parameters can also be specified as step parameters using the global configuration file:

* `artifactType`
* `buildTool`
* `commitVersion`
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

## Example

```groovy
artifactSetVersion script: this, buildTool: 'maven'
```


