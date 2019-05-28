# Project "Piper" User Documentation

An efficient software development process is vital for success in building
business applications on SAP Cloud Platform or SAP on-premise platforms. SAP
addresses this need for efficiency with project "Piper". The goal of project
"Piper" is to substantially ease setting up continuous delivery processes for
the most important SAP technologies by means of Jenkins pipelines.

## What you get

Project "Piper" consists of two parts:

* [A shared library][piper-library] containing steps and utilities that are
  required by Jenkins pipelines.
* A set of [Docker images][devops-docker-images] used in the piper library to implement best practices.

The shared library contains all the necessary steps to run our best practice
[Jenkins pipelines][piper-library-pages] described in the Scenarios section or
to run a [pipeline as step][piper-library-scenario].

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

All steps (`vars` and `resources` directory) are intended to be used by Pipelines and are considered API.
All the classes / groovy-scripts contained in the `src` folder are by default not part of
the API and are subjected to change without prior notice. Types and methods annotated with
`@API` are considered to be API, used e.g. from other shared libraries. Changes to those
methods/types needs to be announced, discussed and agreed.

[github]: https://github.com
[piper-library]: https://github.com/SAP/jenkins-library
[devops-docker-images]: https://github.com/SAP/devops-docker-images
[devops-docker-images-issues]:       https://github.com/SAP/devops-docker-images/issues
[devops-docker-images-cxs-guide]:     https://github.com/SAP/devops-docker-images/blob/master/docs/operations/cx-server-operations-guide.md
[piper-library-scenario]: https://sap.github.io/jenkins-library/scenarios/ui5-sap-cp/Readme/
[piper-library-pages]: https://sap.github.io/jenkins-library
[piper-library-pages-plugins]: https://sap.github.io/jenkins-library/jenkins/requiredPlugins
[piper-library-issues]: https://github.com/SAP/jenkins-library/issues
[piper-library-license]: ./LICENSE
[piper-library-contribution]: .github/CONTRIBUTING.md
[jenkins-doc-pipelines]: https://jenkins.io/solutions/pipeline
[jenkins-doc-libraries]: https://jenkins.io/doc/book/pipeline/shared-libraries
[jenkins-doc-steps]: https://jenkins.io/doc/pipeline/steps
[jenkins-plugin-sharedlibs]: https://wiki.jenkins-ci.org/display/JENKINS/Pipeline+Shared+Groovy+Libraries+Plugin
[google-group]: https://groups.google.com/forum/#!forum/project-piper
