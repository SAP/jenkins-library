# Getting Started with the Guided Tour

## Prerequisites

* Linux - A Linux System with at least 4GB memory. All our samples were tested on Ubuntu 16.04. On Microsoft Windows you might face issues.
* Docker - All tests were made on Docker 18.09.6. Install the newest version from [docker.com](https://docs.docker.com/install/)
* Jenkins - Jenkins version 2.60.3 or higher. We recommend to use the CX-Server toolkit.  
* Access to [github.com][github] - The piper library will be downloaded from [github.com].
* Git Repository - The pipeline you are going to setup will process sources from a Git repository you have to specify. 
* SAP Cloud Platform Space - Get access to SAP Cloud Platform Cloud Foundry. If you haven't an account yet, request a SAP CP CF Trial account. After login an organization and space are targeted. The application will be deployed in this organization and space.

## Jenkins

We offer the life-cycle management toolkit around Jenkins named `cx-server` to ease its usage and configuration. Based on docker images, you will get a preconfigured Jenkins and a Nexus based cache. 
Optionally, you can still use your [own Jenkins installation][guidedtour-my-own-jenkins].

### Jenkins by CX-Server

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


## Creating your first Pipeline

1. Provide a repository on GitHub. For the first time, we recommend to use a sample application SAP provides on [github.com][github]. The repository [cloud-cf-helloworld-nodejs][cloud-cf-helloworld-nodejs] contains a simple `nodejs` application we are going to enrich with a pipeline building with MTA means and deploying into the cloud foundry. Optionally, you can use any repository you want, but be aware that the following code snippets might not match your application.
   
   Fork [cloud-cf-helloworld-nodejs][cloud-cf-helloworld-nodejs] into your GitHub organisation. 
   
   
1. Create a pipeline. Select the branch `1_REST_persist_in_Memory` of your [cloud-cf-helloworld-nodejs] fork and create a new file named `Jenkinsfile`. Enter the following code and submit it.
   
   ```
   @Library('piper-lib-os') _
    node() {
    stage('prepare') {
      checkout scm
      setupCommonPipelineEnvironment script:this
      }
    }
   ```
   This pipeline code will simply sync the repository. 
   
   For more information about Pipeline and what a Jenkinsfile is, refer to [Using a Jenkinsfile][jenkins-io-jenkinsfile] sections of the Jenkins User Handbook.
   
   
1. Setup a Jenkins Job for your repository. 
   
   Open the Jenkins UI `http://<jenkins-server-address>:<http-port>` and Click the `New Item` menu. Per default the `cx-server` will start Jenkins on HTTP port `80`. If you are not familiar with Jenkins, refer to the [Jenkins User Documentation][jenkins-io-documentation].

   <p align="center">
   ![Clicke New Item](../images/JenkinsHomeMenu-1.png "Jenkins Home Menu")
   </p>  
   Provide a name for your new item (e.g. My First Pipeline) and select `Pipeline`

   <p align="center">
   ![Create Pipeline Job](../images/JenkinsNewItemPipeline-1.png "Jenkins New Item")
   </p>  

   Scroll to the Pipeline options and choose `Pipeline script from SCM`. Choose `Git` as SCM and edit the URL of your Git repository, like `https://github.com/<your-org>/cloud-cf-helloworld-nodejs`. `Save` the changes. 

   <p align="center">
   ![Create Pipeline Job](../images/JenkinsNewItemPipeline-2.png "Jenkins New Item")
   </p>  
   
   If your repository is protected you may have to provide credentials.

1. Run your Pipeline. From the Job UI click `Build Now`.

## Add a build step

1. Add the following snippet to your Jenkinsfile. 
   
   ```
    stage('build') {
      mtaBuild script: this
    }
   ```
   
   The `mtaBuild`  step will call a build tool to build a multi-target application (MTA). The tool consumes a MTA descriptor containing the metadata of all entities comprising an application or used by it during deployment or runtime, and the dependencies between them. If you are not familiar with MTAs please visit [sap.com][sap]. 
   
1. Create the MTA descriptor `mta.yaml` with the following content.
   
   ```
    _schema-version: 2.1.0
    ID: com.sap.piper.node.hello.world
    version: 1.0.0
    description: A Hello World sample application
    provider: SAP Sample generator
    modules:
      - name: piper.node.hello.world
        type: nodejs
        path: .
   ```
   
1. Configure `mtaBuild`. To configure the step to build a MTA for the Cloud Foundry, open/create `.pipeline/config.yml` in your repository and add the following content. 
   
   ```
    general:
    steps:
      mtaBuild:
        buildTarget: 'CF'
   ```

For additional information about the configuration refer to the [common configuration guide][resources-configuration] and the [MTA build step documentation][resources-step-mtabuild].

1. Commit the changes.

1. Run your Pipeline. From the Job UI click `Build Now`.

## Add a deploy step

1. Add the following snippet to your Jenkinsfile. 
   
   ```
   stage('deploy') {
     def mtarFilePath = commonPipelineEnvironment.getMtarFilePath()
   
    cloudFoundryDeploy( script: this, mtaPath: mtarFilePath)
   }
   ```
   
   The `cloudFoundryDeploy`  step will call the cloud foundry command line client to deploy into the SAP Cloud Platform. with MTAs please visit [sap.com][sap]. 

1. Configure `cloudFoundryDeploy`. To configure the step to deploy into the Cloud Foundry, open/create `.pipeline/config.yml` in your repository and add the following content. 

   ```
    cloudFoundryDeploy:
      deployTool: 'mtaDeployPlugin'
      deployType: 'standard'
      cloudFoundry:
        org: '<your-organisation>'
        space: '<your-space>'
        credentialsId: 'CF_CREDENTIALSID'
   ```

   For additional information about the configuration refer to the [common configuration guide][resources-configuration] and the [Cloud Foundry deploy step documentation][resources-step-cloudFoundryDeploy].

1. Commit the changes.

1. Run your Pipeline. From the Job UI click `Build Now`.

## Complete Tour  
Your application has been deployed into your SAP CP CF Space. Login and verify the status of the application.
   <p align="center">
   ![Deployed Application](../images/SCPDeployApp-1.png "SAP Cloud Platform")
   </p>  
Click the application name to see the URL of the application. Open the `Route` and add `/users` to the URL. The application will return data.  

If your pipeline fails compare to the final [Jenkinsfile][guidedtour-sample.jenkins], the [config.yml][guidedtour-sample.config] and the [mta.yaml][guidedtour-sample.mta]

## What's Next
This Guided Tour introduced you to the basics of using `Project Piper`. By the concept of Pipeline as Code, Piper respectively Jenkins Pipelines are extremely powerful. While Jenkins Pipelines offer a full set of common programming features, Piper adds SAP specific flavors.

The configuration pattern fosters simple pipelines which can be re-used by multiple applications. Read the documentation of the [configuration][resources-configuration] to understand its principle of inheritance and customization.
 
The `Project Piper`s [Steps][resources-steps] implement the SAP flavors. Have a look into the increasing list of features and visit the different [scenarios][resources-scenarios] to understand how to integrate SAP systems into your pipeline. 


[guidedtour-my-own-jenkins]:         myownjenkins.md
[guidedtour-sample.config]:          samples/cloud-cf-helloworld-nodejs/.pipeline/config.yml
[guidedtour-sample.jenkins]:         samples/cloud-cf-helloworld-nodejs/Jenkinsfile
[guidedtour-sample.mta]:             samples/cloud-cf-helloworld-nodejs/mta.yaml
[resources-configuration]:           configuration.md
[resources-steps]:                   steps
[resources-step-mtabuild]:           steps/mtaBuild.md
[resources-step-cloudFoundryDeploy]: steps/cloudFoundryDeploy.md
[resources-scenarios]:               scenarios
[devops-docker-images]:              https://github.com/SAP/devops-docker-images
[devops-docker-images-issues]:       https://github.com/SAP/devops-docker-images/issues
[cloud-cf-helloworld-nodejs]:  [https://github.com/SAP/cloud-cf-helloworld-nodejs]
[license]:                     LICENSE
[contribution]:                CONTRIBUTING.md
[sap]:                         https://www.sap.com
[github]:                      https://github.com
[jenkins-io-documentation]:    https://jenkins.io/doc/
[jenkins-io-jenkinsfile]:      https://jenkins.io/doc/book/pipeline/jenkinsfile

