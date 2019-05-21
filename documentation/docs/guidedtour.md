# Getting Started with the Guided Tour

##Prerequisites

* Linux - A Linux System with at least 4GB memory. All our samples were tested on Ubuntu 16.04. On Microsoft Windows you might face issues.
* Docker - All tests were made on Docker 18.09.6. Install the newest version from [docker.com](https://docs.docker.com/install/)
* Jenkins - Jenkins version 2.60.3 or higher. We recommend to use the CX-Server toolkit.  
* Access to [github.com][github] - The piper library will be downloaded from [github.com].
* Git Repository - The pipeline you are going to setup will process sources from a Git repository you have to specify. 

##Jenkins

We offer the life-cycle management toolkit around Jenkins named `cx-server` to ease its usage and configuration. Based on docker images, you will get a preconfigured Jenkins and a Nexus based cache. 
Optionally, you can still use your [own Jenkins installation][guidedtour-my-own-jenkins].

###Jenkins by CX-Server

The `cx-server` is a toolkit that is developed to manage the lifecycle of the Jenkins server.
In order to use the toolkit, get the script `cx-server` and its configuration file `server.cfg` by the docker command

```sh
docker run -it --rm -u $(id -u):$(id -g) -v "${PWD}":/cx-server/mount/ ppiper/cx-server-companion:latest init-cx-server
```

Once the files are downloaded into the current directory, you can launch the below command to start the Jenkins server.

```sh
./cx-server start
```

If you would like to customize the Jenkins, [the operations guide](https://github.com/SAP/devops-docker-images/blob/master/docs/operations/cx-server-operations-guide.md) will provide more information on this along with the lifecycle management of the Jenkins. 


# Creating your first Pipeline

1. Copy the following examples into your repository and name it `Jenkinsfile`

        @Library('piper-lib-os') _
        node() {
          stage('prepare') {
              checkout scm
              setupCommonPipelineEnvironment script:this
          }
        }
   The sample will simply sync your repository. 
    
   For more information about Pipeline and what a Jenkinsfile is, refer to the respective Pipeline and [Using a Jenkinsfile][jenkins-io-jenkinsfile] sections of the Jenkins User Handbook.


1. Open the Jenkins UI and Click the New Item menu. Per default the `cx-server` will start Jenkins on HTTP port `80` 

   ![Clicke New Item](../images/JenkinsHomeMenu-1.png "Jenkins Home Menu")

1. Provide a name for your new item (e.g. My First Pipeline) and select `Pipeline`

   ![Create Pipeline Job](../images/JenkinsNewItemPipeline-1.png "Jenkins New Item")

Scroll to the Pipeline options and choose `Pipeline script from SCM`. Choose `Git` as SCM and edit the URL of your Git repository. `Save` the changes.

# Running your Pipeline

# Add a build step

Add the following snippet to your Jenkinsfile. 

          stage('build') {
              mtaBuild script: this
          }

The `mtaBuild`  step will call the MTA build tool to build a multi-target application. If you are not familiar with MTAs please visit [sap.com][sap]. 

To configure the step to build a MTA for the Cloud Foundry, open/create `.pipeline/config.yml` in your repository and add the following content. 

          general:
          steps:
            mtaBuild:
              buildTarget: 'CF'

For additional information about the configuration  refer to the [common configuration guide][resources-configuration] and the [MTA build step documentation][resources-step-mtabuild].

# Pipeline Configuration

# Running your Pipeline

# Additional
## Transport
          stage('prepare') {
              checkout scm
              setupCommonPipelineEnvironment script:this
              checkChangeInDevelopment script: this
          }


# Quick Start Examples


test

[guidedtour-my-own-jenkins]:   myownjenkins.md
[resources-configuration]:     configuration.md
[resources-step-mtabuild]:     steps/mtaBuild.md
[devops-docker-images]:        https://github.com/SAP/devops-docker-images
[devops-docker-images-issues]: https://github.com/SAP/devops-docker-images/issues
[license]:                     LICENSE
[contribution]:                CONTRIBUTING.md
[sap]:                         https://www.sap.com
[github]:                      https://github.com
[jenkins-io-jenkinsfile]:      https://jenkins.io/doc/book/pipeline/jenkinsfile

