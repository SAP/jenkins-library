[![Build Status](https://travis-ci.org/SAP/jenkins-library.svg?branch=master)](https://travis-ci.org/SAP/jenkins-library)
[![Coverage Status](https://coveralls.io/repos/github/SAP/jenkins-library/badge.svg?branch=master)](https://coveralls.io/github/SAP/jenkins-library?branch=master)

# Description

An efficient software development process is vital for success in building
business applications on SAP Cloud Platform or SAP on-premise platforms. SAP
addresses this need for efficiency with project "Piper". The goal of project
"Piper" is to substantially ease setting up continuous deployment processes for
the most important SAP technologies by means of Jenkins pipelines.

Project "Piper" consists of two parts:

 * [A shared library][piper-library] containing steps and utilities that are
   required by Jenkins pipelines.
 * A set of [Jenkins pipelines][piper-pipelines] using the piper library to
   implement best practice processes.

Please follow [this link to our extended library documentation][piper-library-pages].

## What you get

The shared library contains all the necessary steps to run our best practice
[Jenkins pipelines][piper-pipelines].

The best practice pipelines are based on the general concepts of [Jenkins 2.0
Pipelines as Code][jenkins-doc-pipelines].  With that you have the power of the
Jenkins community at hand to optimize your pipelines.

You can run the best practice Jenkins pipelines out of the box, take them as a
starting point for project-specific adaptations or implement your own pipelines
from scratch using the shared library.

## Extensibility

If you consider adding additional capabilities to your `Jenkinsfile`, consult
the [Jenkins Pipeline Steps Reference][jenkins-doc-steps]. There, you get an
overview about steps that are natively supported by Jenkins.

The [Jenkins shared libraries][jenkins-doc-libraries] concept helps you to
extract reusable parts from your pipeline and to keep your pipeline code small
and easy to maintain.

Custom library steps can be added using a custom library according to the
[Jenkins shared libraries][jenkins-doc-libraries] concept instead of adding
groovy coding to the `Jenkinsfile`. Your custom library can coexist next to the
provided pipeline library.

## API

All steps are intended to be used by Pipelines. All the classes / groovy-scripts
contained in the `src` folder are not part of the API and are subjected to change
without prior notice.

# Requirements

 * Java Runtime Environment 8
 * Installation of Jenkins v 2.60.3 or higher running on Linux. We tested with
   debian-stretch.
 * Jenkins Plugins installed as described in the [Required
   Plugin][piper-library-pages-plugins] section.
 * A Jenkins user with administration privileges.
 * The Jenkins instance has access to [github.com][github].

# Download and Installation

To setup the shared library, you need to perform the following steps:

1. Login to your Jenkins instance with administration privileges.
1. Open the system configuration page (*Manage Jenkins > Configure System*).
1. Scroll down to section *Global Pipeline Libraries* and add a new Library by
   clicking the *Add* button.
    1. set *Library Name* to `piper-library-os`
    1. set *Default Version* to the branch or tag you want to consume (e.g.
       `master` or `v0.1`)
    1. set *Retrieval Method* to `Modern SCM`
    1. set *Source Code Management* to `Git`
    1. set *Project Repository* to `https://github.com/SAP/jenkins-library`
1. Save changes

![Library Setup](./documentation/docs/images/setupInJenkins.png)

Now the library is available as `piper-library-os` and can be used in any
`Jenkinsfile` by adding this line:

```
@Library('piper-library-os') _
```

Jenkins will download the library during execution of the `Jenkinsfile`.

# Known Issues

A list of known issues is available on the [GitHub issues page of this
project][piper-library-issues].

# How to obtain support

Feel free to open new issues for feature requests, bugs or general feedback on
the [GitHub issues page of this project][piper-library-issues].

# Contributing

Read and understand our [contribution guidelines][piper-library-contribution]
before opening a pull request.

# [License][piper-library-license]

Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.
This file is licensed under the Apache Software License, v. 2 except as noted
otherwise in the [LICENSE file][piper-library-license]

[github]: https://github.com
[piper-library]: https://github.com/SAP/jenkins-library
[piper-pipelines]: https://github.com/SAP/jenkins-pipelines
[piper-library-pages]: https://sap.github.io/jenkins-library
[piper-library-pages-plugins]: https://sap.github.io/jenkins-library/jenkins/requiredPlugins
[piper-library-issues]: https://github.com/SAP/jenkins-library/issues
[piper-library-license]: ./LICENSE
[piper-library-contribution]: .github/CONTRIBUTING.md
[jenkins-doc-pipelines]: https://jenkins.io/solutions/pipeline
[jenkins-doc-libraries]: https://jenkins.io/doc/book/pipeline/shared-libraries
[jenkins-doc-steps]: https://jenkins.io/doc/pipeline/steps
[jenkins-plugin-sharedlibs]: https://wiki.jenkins-ci.org/display/JENKINS/Pipeline+Shared+Groovy+Libraries+Plugin
