# Extensibility

When using one of the ready-made pipelines project "Piper" provides, the basic idea is to not write custom pipeline code.
The pipelines are centrally maintained, and can be used with a small amount of declarative configuration as documented [here](configuration.md).

For the large majority of _standard_ projects, the features of the ready-made pipelines should be enough to implement [Continuous Delivery](https://martinfowler.com/bliki/ContinuousDelivery.html) in a best-practice compliant way with little effort.
If a feature you need is missing, or you discovered a bug in one of the ready-made pipelines, please see if there is already an [issue in our GitHub repository](https://github.com/SAP/jenkins-library/issues), and open a new one if that is not the case.

In some cases, specialized features might not be desirable for inclusion in the ready-made pipelines.
You can still benefit from the qualities they provide if you can address your requirements via an **Extension**.
Extensions are custom bits of pipeline coding that you can use to implement special requirements.
Before building extensions, please make sure that there is no alternative which works better for you.

Options for extensibility, in the order in which we recommend considering them:

## 1) Extend individual stages

In this option, you use the centrally maintained pipeline, but can change individual stages if required.

To do so, create a file called `<StageName>.groovy` (for example, `Acceptance.groovy` or `lint.groovy`) in `.pipeline/extensions/` in your application's source code repository.

For this, you need to know the technical identifiers for stage names.

* For the general purpose pipeline, you can find them in [the pipeline source file](https://github.com/SAP/jenkins-library/blob/master/vars/piperPipeline.groovy).
* For SAP Cloud SDK Pipeline, you can find them in [this GitHub search query](https://github.com/SAP/cloud-s4-sdk-pipeline-lib/search?q=%22def+stageName+%3D%22).

The centrally maintained pipeline checks if such a file exists and executes it, if present.
A parameter of type `Map` that contains the following keys is passed to the extension:

* `script`: defines the global script environment of the `Jenkinsfile` run. This makes sure that the correct configuration environment can be passed to project "Piper" steps and allows access to for example the `commonPipelineEnvironment`.
* `originalStage`: this will allow you to execute the "original" stage at any place in your script. If omitting a call to `originalStage()` only your code will be executed instead.
* `stageName`: name of the current stage
* `config`: configuration of the stage and general config (including all defaults)

Here is a simple example for such an extension, which you can use as a starting point:

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  //access overall pipeline script object
  echo "Branch: ${params.script.commonPipelineEnvironment.gitBranch}"

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

!!! note "`return this`"
    Don't forget the `return this` which is required at the end of _all_ extension scripts.
    This is due to how Groovy loads scripts internally.

!!! note "Init stage cannot be extended"
    Please note, the `Init` stage among other things also checks out your current repository including your extensions.<br />
    Thus it is not possible to use extensions on this stage.

### Practical example

For a more practical example, you can use extensions in SAP Cloud SDK Pipeline to add custom linters to the pipeline.

A linter is a tool that can check the source code for certain stylistic criteria, and many teams chose to use a linter to ensure a common programming style.

As an example, if you want to use [Checkstyle](https://checkstyle.sourceforge.io/) in your codebase, you might use an extension similar to this one in a file called `.pipeline/extensions/lint.groovy` in your project:

```groovy
def call(Map parameters) {

    parameters.originalStage() // Runs the built in linters

    mavenExecute(
        script: parameters.script,
        flags: '--batch-mode',
        pomPath: 'application/pom.xml',
        m2Path: s4SdkGlobals.m2Directory,
        goals: 'checkstyle:checkstyle',
    )

    recordIssues blameDisabled: true,
        enabledForFailure: true,
        aggregatingResults: false,
        tool: checkStyle()
}

return this
```

This example can be adapted for other linters of your choice.

## 2) Modified ready-made pipeline

This option describes how you can copy and paste one of the centrally maintained pipelines to make changes not possible otherwise.
For example, you can't change the order of stages, change which stages run in parallel or add new stages to a centrally maintained pipeline.
This might be done for an individual project (in the `Jenkinsfile`), or in a separate git repository so it can be used for multiple projects.

### Single project

The default `Jenkinsfile` of centrally maintained pipelines does nothing except for loading the pipeline and running it.
This is convenient, but limits which aspects of the pipeline are modifiable.

If you have one project using the pipeline, the easiest way to do this modification is to copy the pipeline into your `Jenkinsfile`.

The basic structure of your `Jenkinsfile` should be like this:

```groovy
@Library(/* Shared library definition, see ¹ */) _

call script: this

void call(parameters) {
  // Your pipeline code based on our ready-made pipelines
}
```

The actual pipeline code (the `call` method in the listing above) can be found here:

* [General purpose pipeline](https://github.com/SAP/jenkins-library/blob/master/vars/piperPipeline.groovy)
    * ¹ For this pipeline, you need to load this library: `'piper-lib-os@vINSERT_VERSION_HERE'`
* [SAP Cloud SDK Pipeline](https://github.com/SAP/cloud-s4-sdk-pipeline-lib/blob/master/vars/cloudSdkPipeline.groovy)
    * ¹ For this pipeline, you need to load this library: `'s4sdk-pipeline-library@vINSERT_VERSION_HERE'`

For the version identifier, please see the section _How to stay up-to-date_ in this document.

### Multiple projects

If you have multiple projects that share a similar architecture, it might be desirable to share one modified pipeline amongst them.
Similar to what you can do in an individual `Jenkinsfile`, you can copy the pipeline to your own shared library and modify it.

To do this, create a new git repository in your preferred git hosting service.
It must be compliant to [how Jenkins shared libraries are built](https://jenkins.io/doc/book/pipeline/shared-libraries/).
In a nutshell, this means you need a `vars` directory inside which you can place a copy of your preferred pipeline.

A minimal example of such a library could have this directory structure:

```
./vars/myCustomPipeline.groovy
./README.md
```

Where `myCustomPipeline.groovy` contains the modified pipeline code of the [general purpose pipeline](https://github.com/SAP/jenkins-library/blob/master/vars/piperPipeline.groovy) or [SAP Cloud SDK Pipeline](https://github.com/SAP/cloud-s4-sdk-pipeline-lib/blob/master/vars/cloudSdkPipeline.groovy).

!!! note
    Your custom pipeline _must_ be named differently from the other pipelines provided by project "Piper", because Jenkins requires names across multiple libraries to be unique.

This library must be placed in a git repository which is available for Jenkins and must be configured in Jenkins [as documented here](https://jenkins.io/doc/book/pipeline/shared-libraries/#using-libraries).

This screenshot shows an example of the configuration in Jenkins.
Note the name (1) must be the same you use in your `Jenkinsfile`.

![Library Setup](images/customPipelineLib.png "Library Setup")

The `Jenkinsfile` of your individual projects would look similar to this:

```groovy
@Library(['piper-lib-os@vINSERT_VERSION_HERE','my-own-pipeline@vINSERT_VERSION_HERE']) _

myCustomPipeline script: this
```

Be sure to adapt the names and version identifiers accordingly, as described in _How to stay up-to-date_.

### How to stay up-to-date

Regardless of which of the above options you choose, one downside of this approach is that your pipeline will be out of sync with the centrally maintained pipelines at some point in time.
We strongly recommend doing _as little modification as possible_ to fulfil your requirements.
Please be aware that stages may have dependencies on each other.

!!! warning "Don't depend on stage implementation details"
    Your pipeline should treat _stages_ as a black box, the implementation of stages is no published API and may be subject to change at any time.

!!! warning "Beware of breaking changes"
    Please be aware that when using the `master` branch of a library, it might always happen that breaking changes occur.
    We recommend to always fix versions to a released version like in this example: `@Library('my-shared-library@1.0') _`<br />
    Find the most recent release for [jenkins-library](https://github.com/SAP/jenkins-library/releases) and for [SAP Cloud SDK Pipeline](https://github.com/SAP/cloud-s4-sdk-pipeline/releases) on GitHub.
    We do recommend to ["watch" releases for those repositories on GitHub](https://help.github.com/en/github/receiving-notifications-about-activity-on-github/watching-and-unwatching-releases-for-a-repository).

!!! note "When to go with a modified ready-made pipeline"
    This option is right for you, when none of the provided ready-made pipelines serves your purpose, and individual stage extensions don't provide enough flexibility.

### Advanced tips and information

When you consider adding additional capabilities, your first stop should be the [Jenkins Pipeline Steps Reference](https://jenkins.io/doc/pipeline/steps/).
Here you get an overview about what kind of capabilities are already available, and a list of related parameters which you can use to customize the existing implementation.
The provided information should help you to understand and extend the functionality of your pipeline.

## 3) New pipeline from scratch

Since project "Piper" fully builds on [Jenkins Pipelines as Code](https://jenkins.io/doc/book/pipeline-as-code/), you can also go with your own pipeline from scratch in a `Jenkinsfile`.

!!! danger "Decoupling"
    If you go this route you will be decoupled from the innovations provided with project "Piper", unless you re-use for example stages (as indicated above under _2) Modified ready-made pipelines_).

    **We recommend to use this only when none of the other provided options suit your use-case.**
