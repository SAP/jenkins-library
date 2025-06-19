# Project "Piper" general purpose pipeline

The pipeline consists of a sequence of stages where each contains a number of individual steps.

## First step: Pull Request Pipeline

In order to validate pull-requests to your GitHub repository you need to perform two simple steps:

### 1. Create Pipeline configuration

Create a file `.pipeline/config.yml` in your repository (typically in `master` branch) with the following content:

``` YAML
general:
  buildTool: 'npm'
```

!!! note "buildTool"
    Please make sure that you specify the correct build tool.
    Following are currently supported:

    * `docker`
    * `kaniko`
    * `maven`
    * `mta`
    * `npm`

    If your build tool is not in the list you can still use further options as described for [Pull-Request Voting Stage](prvoting.md)

### 2. Create Jenkinsfile

Create a file called `Jenkinsfile` in the root of your repository (typically in `master` branch) with the following content:

```groovy
@Library('piper-lib-os') _

piperPipeline script: this
```

**There is typically no need to further touch this file**

!!! note "Using custom defaults"
    It is possible to overwrite/extend the pipeline defaults with custom defaults.

    ```
    piperPipeline script: this, customDefaults: ['myCustomDefaults.yml']
    ```

    You find more details about the custom defaults in the [configuration section](../configuration.md)

!!! warning "using dedicated versions"
    It is possible to use a fixed version of the library using e.g.

    ```
    @Library('piper-lib-os@v1.222.0') _
    ```

    **Make sure to only use valid git tags as versions!**

## Second step: Prepare pipeline for your main branch

Extend your configuration to also contain git ssh credentials information.

Your `.pipeline/config.yml` should then look like:

``` YAML
general:
  buildTool: 'npm'
  gitSshKeyCredentialsId: 'credentials-id-in-jenkins'
```

!!! note "gitSshKeyCredentialsId"
    The pointer to the Jenkins credentials containing your ssh private key is an important part of the pipeline run.
    The credentials are for example required to push automatic versioning information to your GitHub repository.

## Subsequent steps: Configure individual stages

The stages of the pipeline can be configured individually.
As a general rule of thumb, only stages with an existing configuration are executed.

If no dedicated configuration is required for a step, the precence of relevant files in the repository trigger the step execution.

**This smart and context-aware way of configuration** allows you an iterative approach to configuring the individual steps.

The pipeline comprises following stages:

### Init

This stage takes care that the pipeline is initialized correctly.
It will for example:

* Check out the GitHub repository
* Set up the overall pipeline configuration and perform basic checks
* Identify which pipeline stages to execute based on the configuration and file patterns
* Perform automatic versioning of the software artifact in case the `master` branch pipeline is executed.

You find details about this stage on  [**Init Stage** Details](init.md)

### Pull-Request Voting

This stage is responsible for validating pull-requests, see also above.

You find further details about this stage on the page [**Pull-Request Voting**](prvoting.md).

### Build

In this stage the build of the software artifact is performed.
The build artifact will be `stash`ed for use in subsequent stages. For `Docker` builds the build result will be uploaded to a container registry (as per your configuration).

Afterwards the results of static checks & unit tests are published on the Jenkins.

You find details about this stage on the page [**Build**](build.md).

### Additional Unit Tests

In this stage additional unit-like tests are executed which should not run during the build.

Currently, this stage holds the execution of a Karma runner which allows for

* qUnit tests
* OPA5 (One Page Acceptance tests) for SAPUI5

You find details about this stage on the page [**Additional Unit Tests**](additionalunittests.md).

### Integration

The [Integration stage](integration.md) allows to run test based on maven, npm, or a custom integration test script.
If more flexibility is required, consider using the [stage extension mechanism](../extensibility.md).

You find details about this stage on the page [**Integration**](integration.md).

### Acceptance

In this stage the application/service is typically deployed and automated acceptance tests are executed.

This is to make sure that

* new functionality is tested end-to-end
* there is no end-to-end regression in existing functionality

You find details about this stage on the page [**Acceptance**](acceptance.md).

### Security

This stage can run security checks using Checkmarx, Blackduck Detect, Fortify and WhiteSource.

You find details about this stage on the page [**Security**](security.md).

### Performance

The stage will execute a Gatling test, if the step `gatlingExecuteTests` is configured.

You find details about this stage on the page [**Performance**](performance.md).

### Compliance

The stage will execute a SonarQube scan, if the step `sonarExecuteSan` is configured.

You find details about this stage on the page [**Compliance**](compliance.md).

### Confirm

The [Confirm stage](confirm.md), if executed, stops the pipeline execution and asks for manual confirmation before proceeding to the stages _Promote_ and _Release_.

### Promote

This stage is responsible to promote build artifacts to an artifact repository / container registry where they can be used from production deployments.

You find details about this stage on the page [**Promote**](promote.md).

### Release

This stage is responsible to release/deploy artifacts into your productive landscape.

You find details about this stage on the page [**Release**](release.md).
