# Artifact Deployment to Nexus

## Status

Accepted

## Context

Deploying artifacts to a Nexus Repository Manager needs to maintain `maven-metadata.xml` files in order to be compatible with tools consuming the artifacts. 

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
- A list of parameters has to be generated before using the plugin, including `artifactId` and `version`, which is the same case as the `Uploading artifacts manually`. For maven projects, the parameters can be obtained using the `evaluate` goal of the `maven-help-plugin`. There is however a performance impact, since a maven command line has to be executed for each parameter, multiplied by the number of modules. This is not a problem for `Maven lifecycle phase : deploy`.
- Credential info has to be stored in a `settings.xml`, which introduces additional implementation. Credentials can be passed via environment variables.


#### Maven lifecycle phase: deploy
By default, the maven lifecycle phase `deploy` binds to the goal `deploy:deploy` of the `Apache Maven Deploy Plugin`.
##### :+1:
- Same as the `Apache Maven Deploy Plugin`
- You don't have to obtain and pass the parameters as for `Apache Maven Deploy Plugin`, because `package` phase is executed implicitly and makes the parameters ready before `deploy` phase.
##### :-1:
- Same case as the `Apache Maven Deploy Plugin` for handling credentials.
- Cannot be used for non-Maven projects (i.e. MTA)
- As a maven phase, a list of phases is triggered implicitly before this phase, including `compile`, `test` and `package`.
To follow the build-once principle, all these phases have to be skipped.
However, it's not possible to skip some of the maven goals binding to certain phases.
For example, if the `<packaging>` tag of the `pom.xml` is set to `jar`, then the `jar:jar` goal of the [`Apache Maven JAR Plugin`](https://maven.apache.org/plugins/maven-jar-plugin/) is bound to `package` phase.
Unfortunately, however, `Apache Maven JAR Plugin` does not provide an option to skip the the `jar:jar` goal. There may be a [solution](https://stackoverflow.com/questions/47673545/how-to-skip-jar-deploy-in-maven-and-deploy-the-assembly-only), but it seems to require modifying the pom and could also be different depending on the used packaging.  
**This is the main reason why we cannot use this option.**


#### Uploading artifacts manually
##### :+1:
- Without the pain of handling the credentials, which was mentioned above in `Apache Maven Deploy Plugin` section.
- Gives full control over the implementation. 
##### :-1:
- Same as the `Apache Maven Deploy Plugin`. A list of parameters has to be prepared.
- Introduces complexity for maintaining maven-metadata.xml. For example there is a great difference between "release" and "snapshot" deployments. The later have a build number and another directory structure on the nexus (arbitrary number of builds per version, with metadata for each build and for the version). 
- Has the greatest maintenance-overhead.

### Decision
`Apache Maven Deploy Plugin` is chosen, because:
- `Maven lifecycle phase: deploy` does not meet our build-once principle.
- Credential handling is not very complex to implement.
- Has the fine-grained control needed over which artifacts are deployed.
- Maintains maven-metadata.xml correctly for various types of deployments.