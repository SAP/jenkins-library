# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

none

## ${docGenParameters}

Details:

The step is stashing files before and after the build. This is due to the fact, that some of the code that needs to be stashed, is generated during the build (TypeScript for NPM).

| stash name | mandatory | prerequisite | pattern |
|---|---|---|---|
|buildDescriptor|no| |includes: `**/pom.xml, **/.mvn/**, **/assembly.xml, **/.swagger-codegen-ignore, **/package.json, **/requirements.txt, **/setup.py, **/whitesource_config.py, **/mta*.y*ml, **/.npmrc, **/whitesource.*.json, **/whitesource-fs-agent.config, Dockerfile, **/VERSION, **/version.txt, **/Gopkg.*, **/dub.json, **/dub.sdl, **/build.sbt, **/sbtDescriptor.json, **/project/*`<br /> excludes: `**/node_modules/**/package.json`|
|checkmarx|no|Checkmarx is enabled|includes: `**/*.js, **/*.scala, **/*.go, **/*.d, **/*.di`<br /> excludes: `**/*.mockserver.js, node_modules/**/*.js`|
|classFiles|no| |includes: `**/target/classes/**/*.class, **/target/test-classes/**/*.class` <br />excludes: `''`|
|deployDescriptor|no| |includes: `**/manifest*.y*ml, **/*.mtaext.y*ml, **/*.mtaext, **/xs-app.json, helm/**, *.y*ml`<br />exclude: `''`|
|git|no| |includes: `**/gitmetadata/**`<br />exludes: `''`|
|opensourceConfiguration|no| |includes: `**/srcclr.yml, **/vulas-custom.properties, **/.nsprc, **/.retireignore, **/.retireignore.json, **/.snyk`<br />excludes: `''`|
|pipelineConfigAndTests|no| |includes: `.pipeline/*.*`<br />excludes: `''`|
|securityDescriptor|no| |includes: `**/xs-security.json`<br />exludes: `''`|
|sonar|no| |includes: `**/jacoco*.exec, **/sonar-project.properties`<br />exludes: `''`|
|tests|no| |includes: `**/pom.xml, **/*.json, **/*.xml, **/src/**, **/node_modules/**, **/specs/**, **/env/**, **/*.js`<br />excludes: `''`|

!!! note "Overwriting default stashing behavior"
    It is possible to overwrite the default behavior of the stashes using the parameters `stashIncludes` and `stashExcludes` , e.g.

    * `stashIncludes: [buildDescriptor: '**/mybuild.yml]`
    * `stashExcludes: [tests: '**/NOTRELEVANT.*]`

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Explanation of pipeline step

Usage of pipeline step:

```groovy
pipelineStashFiles script: this {
  mavenExecute script: this, ...
}
```
