# Project "Piper" Overview

An efficient software development process is vital for success in building business applications on SAP Cloud Platform or SAP on-premise platforms.
SAP addresses this need for efficiency with project "Piper". The goal of project "Piper" is to substantially ease setting up continuous deployment processes for the most important SAP technologies by means of Jenkins pipelines.

Project "Piper" consists of two parts:

 * A [shared library][piper-library] containing steps and utilities that are required by Jenkins pipelines.
 * A set of [Jenkins pipelines][piper-pipelines] using the piper library to implement best practice processes as code.

## What you get

The shared library contains all the necessary steps to run our best practice [Jenkins pipelines][piper-pipelines].

!!! note "Jenkins 2.0 Pipelines as Code"
    The best practice pipelines are based on the general concepts of [Jenkins 2.0 Pipelines as Code][jenkins-doc-pipelines].
    With that you have the power of the Jenkins community at hand to optimize your pipelines.

You can run the best practice Jenkins pipelines out of the box, take them as a starting point for project-specific adaptations or implement your own pipelines from scratch using the shared library.

## Installation

Prerequisites:

 * Installation of Jenkins v 2.60.3 or higher running on Linux. We tested with debian-stretch.
 * Jenkins Plugins installed as described in the [Required Plugin](jenkins/requiredPlugins) section.
 * A Jenkins user with administration privileges.
 * The Jenkins instance has access to [github.com](https://github.com).

To setup the shared library, you need to perform the following steps:

1. Login to your Jenkins instance with administration privileges.
1. Open the system configuration page (*Manage Jenkins > Configure System*).
1. Scroll down to section *Global Pipeline Libraries* and add a new Library by clicking the *Add* button.
    1. set *Library Name* to `piper-library-os`
    1. set *Default Version* to the branch or tag you want to consume (e.g. `master` or `v0.1`)
    1. set *Retrieval Method* to `Modern SCM`
    1. set *Source Code Management* to `Git`
    1. set *Project Repository* to `https://github.com/SAP/jenkins-library`
1. Save changes

![Library Setup](images/setupInJenkins.png)

Now the library is available as `piper-library-os` and can be used in any `Jenkinsfile` by adding this line:

```
@Library('piper-library-os') _
```

## Extensibility

If you consider adding additional capabilities to your `Jenkinsfile`, consult the [Jenkins Pipeline Steps Reference][jenkins-doc-steps].
There, you get an overview about steps that are natively supported by Jenkins.

The [Jenkins shared libraries][jenkins-doc-libraries] concept helps you to extract reusable parts from your pipeline and to keep your pipeline code small and easy to maintain.

!!! tip
    If you consider adding custom library steps you can do so using a custom library according to the [Jenkins shared libraries][jenkins-doc-libraries] concept instead of adding groovy coding to the `Jenkinsfile`.
    Your custom library can coexist next to the provided pipeline library.

## Community & Support

In the [GitHub repository of the shared library][piper-library] you can find a list of GitHub issues for known bugs or planned future improvements.
Feel free to open new issues for feature requests, bugs or general feedback.

[piper-library]: https://github.com/SAP/jenkins-library
[piper-pipelines]: https://github.com/SAP/jenkins-pipelines
[jenkins-doc-pipelines]: https://jenkins.io/solutions/pipeline
[jenkins-doc-libraries]: https://jenkins.io/doc/book/pipeline/shared-libraries
[jenkins-doc-steps]: https://jenkins.io/doc/pipeline/steps
[jenkins-plugin-sharedlibs]: https://wiki.jenkins-ci.org/display/JENKINS/Pipeline+Shared+Groovy+Libraries+Plugin
