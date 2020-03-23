# Artifact Deployment to Nexus

## Status

Accepted

## Context

The nexusUpload step shall upload (deploy) build artifacts to a Nexus Repository Manager. Nexus version 2 and 3 need to be supported.
Per module, there can be an artifact and multiple sub-artifacts, which need to be deployed as a unit, optionally together with the project descriptor file (i.e. pom.xml or mta.yaml).
A Nexus contains repositories of different type. For example, a "release" repository does not allow updating existing artifacts, while a "snapshot" repository allows for multiple builds of the same snapshot version, with the notion of a "latest" build.
Depending on the type of repository, a certain directory layout has to be obeyed, and `maven-metadata.xml` files have to be maintained in order to be compatible with tools consuming the artifacts. The Nexus itself may also have mechanisms in place, for example to automatically purge old builds in snapshot releases.
All this makes it important to make compatible deployments.

### Alternatives

* [Apache Maven Deploy Plugin](http://maven.apache.org/plugins/maven-deploy-plugin/)
* Maven lifecycle phase : deploy
* Uploading artifacts manually

### Pros and Cons

#### Apache Maven Deploy Plugin (deploy:deploy-file)

For this option, we only consider the goal `deploy:deploy-file`.

##### :+1:

- Official maven plugin for deployment, which is perfect if you only care whether the artifacts are deployed correctly.

##### :-1:

- Knowledge about which artifacts to deploy has to be obtained manually.
- A list of parameters has to be generated before using the plugin, including `artifactId` and `version`, which is the same case as the `Uploading artifacts manually`. For maven projects, the parameters can be obtained using the `evaluate` goal of the `maven-help-plugin`. There is however a performance impact, since a maven command line has to be executed for each parameter, multiplied by the number of modules. This is not a problem for `Maven lifecycle phase : deploy`.
- Credential info has to be stored in a `settings.xml`, which introduces additional implementation. Credentials can be passed via environment variables.

#### Maven lifecycle phase: deploy

By default, the maven lifecycle phase `deploy` binds to the goal `deploy:deploy` of the `Apache Maven Deploy Plugin`.

##### :+1:

- Same as the `Apache Maven Deploy Plugin`
- You don't have to obtain and pass the parameters as for `Apache Maven Deploy Plugin`, because `package` phase is executed implicitly and makes the parameters ready before `deploy` phase.
- Supports multi-module Maven projects and any project structure.

##### :-1:

- Same case as the `Apache Maven Deploy Plugin` for handling credentials.
- Cannot be used for non-Maven projects (i.e. MTA)
- As a maven phase, a list of phases is triggered implicitly before this phase, including `compile`, `test` and `package`.
To follow the build-once principle, all these phases have to be skipped.
However, it's not possible to skip some of the maven goals binding to certain phases.
For example, if the `<packaging>` tag of the `pom.xml` is set to `jar`, then the `jar:jar` goal of the [`Apache Maven JAR Plugin`](https://maven.apache.org/plugins/maven-jar-plugin/) is bound to the `package` phase.
Unfortunately, however, `Apache Maven JAR Plugin` does not provide an option to skip the the `jar:jar` goal. There may be a [solution](https://stackoverflow.com/questions/47673545/how-to-skip-jar-deploy-in-maven-and-deploy-the-assembly-only), but it seems to require modifying the pom and could also be different depending on the used packaging.
**This is the main reason why we cannot use this option.**

#### Uploading artifacts manually

Files can be uploaded to the Nexus by simple HTTP PUT requests, using basic authentication if necessary. Meta-data files have to be downloaded, updated and re-uploaded after successful upload of the artifacts.

##### :+1:

- Without the pain of handling the credentials, which was mentioned above in `Apache Maven Deploy Plugin` section.
- Gives full control over the implementation.

##### :-1:

- Same as the `Apache Maven Deploy Plugin`. Knowledge about which artifacts to deploy has to be obtained manually.
- Same as the `Apache Maven Deploy Plugin`. A list of parameters has to be prepared.
- Introduces complexity for maintaining maven-metadata.xml. For example there is a great difference between "release" and "snapshot" deployments. The later have a build number and another directory structure on the nexus (arbitrary number of builds per version, with metadata for each build and for the version).
- Has the greatest maintenance-overhead.

### Decision

`Apache Maven Deploy Plugin` is chosen, because:

- `Maven lifecycle phase: deploy` conflicts with the build-once principle.
- Credentials handling is not very complex to implement.
- Has the fine-grained control needed over which artifacts are deployed.
- Maintains maven-metadata.xml correctly for various types of deployments.
