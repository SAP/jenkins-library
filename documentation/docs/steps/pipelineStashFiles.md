# pipelineStashFiles

## Description
This step stashes files that are needed in other build steps (on other nodes).

## Prerequsites
none

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | no | empty `commonPipelineEnvironment` |  |
| runCheckmarx | no | false |  |
| runOpaTests | no | false |  |
| stashIncludes | no | not set |  |
| stashExcludes | no | not set |  |

Details:

The step is stashing files before and after the build. This is due to the fact, that some of the code that needs to be stashed, is generated during the build (TypeScript for NPM).

| stash name | mandatory | prerequisite | pattern |
|---|---|---|---|
|buildDescriptor|no| |includes: `**/pom.xml, **/.mvn/**, **/assembly.xml, **/.swagger-codegen-ignore, **/package.json, **/requirements.txt, **/setup.py, **/whitesource_config.py, **/mta*.y*ml, **/.npmrc, **/whitesource.*.json, **/whitesource-fs-agent.config, .xmake.cfg, Dockerfile, **/VERSION, **/version.txt, **/build.sbt, **/sbtDescriptor.json, **/project/*`<br /> excludes: `**/node_modules/**/package.json`|
|checkmarx|no|Checkmarx is enabled|includes: `**/*.js, **/*.scala, **/*.go`<br /> excludes: `**/*.mockserver.js, node_modules/**/*.js`|
|classFiles|no| |includes: `**/target/classes/**/*.class, **/target/test-classes/**/*.class` <br />excludes: `''`|
|deployDescriptor|no| |includes: `**/manifest*.y*ml, **/*.mtaext.y*ml, **/*.mtaext, **/xs-app.json, helm/**, *.y*ml`<br />exclude: `''`|
|git|no| |includes: `**/gitmetadata/**`<br />exludes: `''`|
|opa5|no|OPA5 is enabled|includes: `**/*.*`<br />excludes: `''`|
|opensourceConfiguration|no| |includes: `**/srcclr.yml, **/vulas-custom.properties, **/.nsprc, **/.retireignore, **/.retireignore.json, **/.snyk`<br />excludes: `''`|
|pipelineConfigAndTests|no| |includes: `.pipeline/*.*`<br />excludes: `''`|
|securityDescriptor|no| |includes: `**/xs-security.json`<br />exludes: `''`|
|sonar|no| |includes: `**/jacoco*.exec, **/sonar-project.properties`<br />exludes: `''`|
|tests|no| |includes: `**/pom.xml, **/*.json, **/*.xml, **/src/**, **/node_modules/**, **/specs/**, **/env/**, **/*.js`<br />excludes: `''`|

!!! note "Overwriting default stashing behavior"
    It is possible to overwrite the default behavior of the stashes using the parameters `stashIncludes` and `stashExcludes` , e.g.

    * `stashIncludes: [buildDescriptor: '**/mybuild.yml]`
    * `stashExcludes: [tests: '**/NOTRELEVANT.*]`

## Step configuration
The following parameters can also be specified as step parameters using the global configuration file:

* runOpaTests
* runCheckmarx
* stashExcludes
* stashIncludes


## Explanation of pipeline step

Usage of pipeline step:

```groovy
pipelineStashFiles script: this {
  mavenExecute script: this, ...
}
```

Available parameters:

