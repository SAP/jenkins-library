# Project "Piper" User Documentation

Continuous delivery is a method to develop software with short feedback cycles.
It is applicable to projects both for SAP Cloud Platform and SAP on-premise platforms.
SAP implements tooling for continuous delivery in project "Piper".
The goal of project "Piper" is to substantially ease setting up continuous delivery in your project using SAP technologies.

## What you get

To get you started quickly, project "Piper" offers you the following artifacts:

* A set of ready-made Continuous Delivery pipelines for direct use in your project
    * [ABAP Environment Pipeline](pipelines/abapEnvironment/introduction/)
    * [General Purpose Pipeline](stages/introduction/)
    * [SAP Cloud SDK Pipeline][cloud-sdk-pipeline]
* [A shared library][piper-library] that contains reusable step implementations, which enable you to customize our preconfigured pipelines, or to even build your own customized ones
* A standalone [command line utility](cli) for Linux and a [GitHub Action](https://github.com/SAP/project-piper-action)
    * Note: This version is still in early development. Feel free to use it and [provide feedback](https://github.com/SAP/jenkins-library/issues), but don't expect all the features of the Jenkins library
* A set of [Docker images][devops-docker-images] to setup a CI/CD environment in minutes using sophisticated life-cycle management

To find out which offering is right for you, we recommend to look at the ready-made pipelines first.
In many cases, they should satisfy your requirements, and if this is the case, you don't need to build your own pipeline.

### The best-practice way: Ready-made pipelines

**Are you building a standalone SAP Cloud Platform application?<br>**
Then continue reading about our [general purpose pipeline](stages/introduction/), which supports various technologies and programming languages.

**Are you building an application with the SAP Cloud SDK and/or SAP Cloud Application Programming Model?<br>**
Then we can offer you a [pipeline specifically tailored to SAP Cloud SDK and SAP Cloud Application Programming Model applications][cloud-sdk-pipeline]

### The do-it-yourself way: Build with Library

The shared library contains building blocks for your own pipeline, following our best practice Jenkins pipelines described in the Scenarios section.

The best practice pipelines are based on the general concepts of [Pipelines as Code, as introduced in Jenkins 2][jenkins-doc-pipelines].
With that you have the power of the Jenkins community at hand to optimize your pipelines.

You can run the best practice Jenkins pipelines out of the box, take them as a
starting point for project-specific adaptations or implement your own pipelines
from scratch using the shared library.

For an example, you might want to check out our ["Build and Deploy SAPUI5 or SAP Fiori Applications on SAP Cloud Platform with Jenkins" scenario][piper-library-scenario].

#### Extensibility

For the vast majority of _standard_ projects, the features of the ready-made pipelines should be enough to implement [Continuous Delivery](https://martinfowler.com/bliki/ContinuousDelivery.html) with little effort in a best-practice compliant way.
If you require more flexibility, our documentation on [Extensibility][piper-doc-extensibility] discusses available options.

#### API

All steps (`vars` and `resources` directory) are intended to be used by Pipelines and are considered API.
All the classes / groovy-scripts contained in the `src` folder are by default not part of
the API and are subjected to change without prior notice. Types and methods annotated with
`@API` are considered to be API, used e.g. from other shared libraries. Changes to those
methods/types needs to be announced, discussed and agreed.


[github]: https://github.com
[piper-library]: https://github.com/SAP/jenkins-library
[cloud-sdk-pipeline]: pipelines/cloud-sdk/introduction/
[devops-docker-images]: https://github.com/SAP/devops-docker-images
[devops-docker-images-issues]:       https://github.com/SAP/devops-docker-images/issues
[devops-docker-images-cxs-guide]:     https://github.com/SAP/devops-docker-images/blob/master/docs/operations/cx-server-operations-guide.md
[piper-library-scenario]: scenarios/ui5-sap-cp/Readme/
[piper-doc-extensibility]: extensibility
[piper-library-pages-plugins]: requiredPlugins
[piper-library-issues]: https://github.com/SAP/jenkins-library/issues
[piper-library-license]: ./LICENSE
[piper-library-contribution]: .github/CONTRIBUTING.md
[jenkins-doc-pipelines]: https://jenkins.io/solutions/pipeline
[jenkins-doc-libraries]: https://jenkins.io/doc/book/pipeline/shared-libraries
[jenkins-doc-steps]: https://jenkins.io/doc/pipeline/steps
[jenkins-plugin-sharedlibs]: https://wiki.jenkins-ci.org/display/JENKINS/Pipeline+Shared+Groovy+Libraries+Plugin
[google-group]: https://groups.google.com/forum/#!forum/project-piper
