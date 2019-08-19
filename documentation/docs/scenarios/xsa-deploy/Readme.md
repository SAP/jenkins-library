# Build and Deploy SAP Fiori Applications on SAP HANA XS Advanced

Build an application based on SAPUI5 or SAP Fiori with Jenkins and deploy the build result into an SAP Cloud Platform account in the Neo environment.

## Prerequisites

* TODO: do we have a general description how to setup docker. Doesn't make sense to describe such general setups on the level of each scenario.
* You have installed the Java Runtime Environment 8. TODO: depends on docker setup.
* You have installed Jenkins 2.60.3 or higher. TODO: not required in case CX Jenkins is used
* You have set up Project “Piper”. See [README](https://github.com/SAP/jenkins-library/blob/master/README.md). TODO: not needed when CX Jenkins is used. Required for plain Jenkins use cases.
* You have installed the Multi-Target Application (MTA) Archive Builder 1.0.6 or newer. See [SAP Development Tools](https://tools.hana.ondemand.com/#cloud). TODO: Obsolete with docker
* You have installed Node.js including node and npm. See [Node.js](https://nodejs.org/en/download/). TODO: obsolete when using docker
* TOOD: Prerequisite is that the artifacts we compile agaist are available either on Service Market Place (next TODO: explain how they can be imported into the build env), or via public maven repo

### Project Prerequisites

This scenario requires additional files in your project and in the execution environment on your Jenkins instance.

On the project level, provide and adjust the following template:

| File Name | Description | Position |
|-----|-----|-----|
| [`mta.yaml`](https://github.com/SAP/jenkins-library/blob/master/documentation/docs/scenarios/xsa-deploy/files/mta.yaml) | This file controls the behavior of the MTA toolset. | Place the `mta.yaml` file in your application root folder and adjust the values in brackets with your data. |

## Context

This scenario combines various different steps to create a complete pipeline.

In this scenario, we want to show how to build a Multitarget Application (MTA) and deploy the build result into an on-prem SAP HANA XS advances system. This document comprises the [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/) and the [xsDeploy](https://sap.github.io/jenkins-library/steps/xsDeploy/) steps.

![This pipeline in Jenkins Blue Ocean](images/pipeline.jpg)
###### Screenshot: Build and Deploy Process in Jenkins

## Example

### Jenkinsfile

Following the convention for pipeline definitions, use a `Jenkinsfile`, which resides in the root directory of your development sources.

TODO: here we describe the straight-forward case. There is also a blue-green mode. How should we handle this

```groovy
@Library('piper-library-os') _

pipeline {

    agent any

    stages {
        stage("prepare") {
            steps {
                deleteDir()
                checkout scm
                setupCommonPipelineEnvironment script: this
            }
        }
        stage('build') {
            steps {
                mtaBuild script: this
            }
        }
        stage('deploy') {
            steps {
                xsDeploy script: this
            }
        }
    }
}
```

### Configuration (`.pipeline/config.yml`)

This is a basic configuration example, which is also located in the sources of the project.

```yaml
steps:
  mtaBuild:
    buildTarget: 'XSA'
  xsDeploy:
    apiUrl: '<API_URL>' # e.g. 'https://example.org:30030'
    # credentialsId: 'XS' omitted, 'XS' is the default
    docker:
      dockerImage: '<ID_OF_THE_DOCKER_IMAGE' # for legal reasons no docker image is provided.
      # dockerPullImage: true # default: 'false'. Needs to be set to 'true' in case the image is served from a docker registry
    loginOpts: '' # during setup for non-productive builds we might set here. '--skip-ssl-validation'
    org: '<ORG_NAME>'
    space: '<SPACE>'

```

#### Configuration for the MTA Build

| Parameter        | Description    |
| -----------------|----------------|
| `buildTarget`    | The target platform to which the mtar can be deployed. In this case we need  `XSA` |

#### Configuration for the Deployment to XSA 

| Parameter          | Description |
| -------------------|-------------|
| `credentialsId` | The Jenkins credentials that contain user and password required for the deployment on SAP Cloud Platform.|
| `mode`          | DeployMode. TODO: we need to provide the details here
| `org`           |  The org TODO: we need to provide the details here |
| `space`           | The space TODO: we need to provide the details here |

### Parameters

For the detailed description of the relevant parameters, see:

* [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/)
* [xsDeploy](https://sap.github.io/jenkins-library/steps/xsDeploy/)
