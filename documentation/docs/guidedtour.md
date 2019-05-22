# Getting Started with Project "Piper"

Follow this guided tour to become familiar with the basics of using project "Piper". 


## Prerequisites

* You have installed a Linux system with at least 4 GB memory. **Note:** We have tested our samples on Ubuntu 16.04. On Microsoft Windows, you might face some issues.
* You have installed the newest version of Docker. See [Docker Community Edition](https://docs.docker.com/install/). **Note:** we have tested on Docker 18.09.6.
* You have installed Jenkins 2.60.3 or higher. **Recommendation:** We recommend to use the `cx-server` toolkit. See **(Optional) Install the `cx-server` Toolkit for Jenkins**.
* You have a GitHub account. See [Signing up for a new GitHub account](https://help.github.com/en/articles/signing-up-for-a-new-github-account).
* You have access to a repository on GitHub. See [Creating a repository on GitHub](https://help.github.com/en/articles/creating-a-repository-on-github).
* You have an account and space in the Cloud Foundry environment on SAP Cloud Platform. See [Get Started with a Trial Account: Workflow in the Cloud Foundry Environment](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/e50ab7b423f04a8db301d7678946626e.html).

## (Optional) Install the `cx-server` Toolkit for Jenkins

`cx-server`is a lifecycle management toolkit that provides Docker images with a preconfigured Jenkins and a Nexus-based cache to facilitate the configuration and usage of Jenkins.

To use the toolkit, get the `cx-server` script and its configuration file `server.cfg` by using the following command:

```sh
docker run -it --rm -u $(id -u):$(id -g) -v "${PWD}":/cx-server/mount/ ppiper/cx-server-companion:latest init-cx-server
```

When the files are downloaded into the current directory, launch the Jenkins server by using the following command:


```sh
./cx-server start
```

For more information on the Jenkins lifecycle management and how to customize your Jenkins, have a look at the [Operations Guide for Cx Server](https://github.com/SAP/devops-docker-images/blob/master/docs/operations/cx-server-operations-guide.md). 


## Create Your First Pipeline

For the beginning, we recommend using an SAP sample application. The repository [cloud-cf-helloworld-nodejs][cloud-cf-helloworld-nodejs] contains a simple `nodejs` application that can be enriched with a pipeline built with MTA and deployed into the Cloud Foundry environment.

1. Fork the [cloud-cf-helloworld-nodejs][cloud-cf-helloworld-nodejs] repository into your GitHub organization.

2. Select the `1_REST_persist_in_Memory` branch of your [cloud-cf-helloworld-nodejs] fork and in it, create a new file with the name `Jenkinsfile`.

3. To synchronize the repository, enter the following code lines into your `Jenkinsfile`: 

   ```
   @Library('piper-lib-os') _
    node() {
    stage('prepare') {
      checkout scm
      setupCommonPipelineEnvironment script:this
      }
    }
   ```
   For more information about Jenkinsfiles and pipelines, see [Using a Jenkinsfile][jenkins-io-jenkinsfile].
   
4. To set up a Jenkins job for your repository, open the Jenkins UI under `http://<jenkins-server-address>:<http-port>` and choose **New Item**. Per default, the `cx-server` starts Jenkins on HTTP port `80`. For more information, see the [Jenkins User Documentation][jenkins-io-documentation].
   <p align="center">
   ![Clicke New Item](../images/JenkinsHomeMenu-1.png "Jenkins Home Menu")
   </p>  
5. Provide a name for your new item (for example, *My First Pipeline*) and select **Pipeline**.

   <p align="center">
   ![Create Pipeline Job](../images/JenkinsNewItemPipeline-1.png "Jenkins New Item")
   </p>  

6. For **Definition** in the **Pipeline** options, choose **Pipeline script from SCM**. 

7. For **SCM**, choose **Git**.

8. For **Repository URL** in the **Repositories** section, enter the URL of your Git repository, for example `https://github.com/<your-org>/cloud-cf-helloworld-nodejs`. **Note:** If your repository is protected, you must provide your credentials in the **Credentials** section.

   <p align="center">
   ![Create Pipeline Job](../images/JenkinsNewItemPipeline-2.png "Jenkins New Item")
   </p>  

8. Choose **Save**. 

9. To run your pipeline, choose **Build Now** in the job UI.


## Add a Build Step

1. In your `Jenkinsfile`, add the following code snippet: 
   ```
    stage('build') {
      mtaBuild script: this
    }
   ```
   **Result:** The `mtaBuild` step calls a build tool to build a multi-target application (MTA). The tool consumes an MTA descriptor that contains the metadata of all entities which comprise an application or are used by one during deployment or runtime, and the dependencies between them. For more information about MTAs, see [sap.com][sap]. 
   
2. Create an MTA descriptor with the name `mta.yaml`, which contains the following code:

   
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
   
3. To configure the step to build an MTA for the Cloud Foundry environment, in your repository, open or create the `.pipeline/config.yml` and add the following content: 
   
   ```
    general:
    steps:
      mtaBuild:
        buildTarget: 'CF'
   ```

   For additional information about the configuration, have a look at the [Common Configuration Guide][resources-configuration] and the [MTA build step documentation][resources-step-mtabuild].

4. Commit your changes.

5. To run your pipeline, choose **Build Now** in the job UI.

## Add a Deploy Step

1.  In your `Jenkinsfile`, add the following code snippet:
   
   ```
   stage('deploy') {
     def mtarFilePath = commonPipelineEnvironment.getMtarFilePath()
     cloudFoundryDeploy( script: this, mtaPath: mtarFilePath)
   }
   ```
   
   **Result:** The `cloudFoundryDeploy`  step calls the Cloud Foundry command line client to deploy into SAP Cloud Platform.

2. To configure the step to deploy into the Cloud Foundry environment, in your repository, open or create the `.pipeline/config.yml` and add the following content:

   ```
    cloudFoundryDeploy:
      deployTool: 'mtaDeployPlugin'
      deployType: 'standard'
      cloudFoundry:
        org: '<your-organisation>'
        space: '<your-space>'
        credentialsId: 'CF_CREDENTIALSID'
   ```
   The key `CF_CREDENTIALSID` refers to a user-password credential you must create in Jenkins: In Jenkins, choose **Credentials** from the main menu and add a **Username with Password** entry.
   
   <p align="center">
   ![Add Credentials](../images/JenkinsCredentials-1.png "Add Credentials")
   </p>  
   
   For more information about the configuration, see the [Common Configuration Guide][resources-configuration] and [cloudFoundryDeploy][resources-step-cloudFoundryDeploy].

3. Commit your changes.

4. To run your pipeline, choose **Build Now** in the job UI.

## Complete the Guided Tour  

Your application has been deployed into your space in the Cloud Foundry space on SAP Cloud Platform. Logon to SAP Cloud Platform and verify the status of your application.
   
   <p align="center">
   ![Deployed Application](../images/SCPDeployApp-1.png "SAP Cloud Platform")
   </p>  
   
To view the URL of your application, choose the application name. Open the **Route** and add `/users` to the URL. **Result:** The application returns data.  

If your pipeline fails, compare it to the final [Jenkinsfile][guidedtour-sample.jenkins], the [config.yml][guidedtour-sample.config], and the [mta.yaml][guidedtour-sample.mta].

## What's Next

You are now familiar with the basics of using project "Piper". Through the concept of pipeline as code, project "Piper" and Jenkins pipelines are extremely powerful. While Jenkins pipelines offer a full set of common programming features, project "Piper" adds SAP-specific flavors. Have a look at the increasing list of features you can implement through the project "Piper" [steps][resources-steps] and see the different [scenarios][resources-scenarios] to understand how to integrate SAP systems into your pipeline.

The configuration pattern supports simple pipelines that can be reused by multiple applications. To understand the principles of inheritance and customization, have a look at the the [configuration][resources-configuration] documentation.
 

[guidedtour-my-own-jenkins]:         myownjenkins.md
[guidedtour-sample.config]:          samples/cloud-cf-helloworld-nodejs/.pipeline/config.yml
[guidedtour-sample.jenkins]:         samples/cloud-cf-helloworld-nodejs/Jenkinsfile
[guidedtour-sample.mta]:             samples/cloud-cf-helloworld-nodejs/mta.yaml
[resources-configuration]:           configuration.md
[resources-steps]:                   steps
[resources-step-mtabuild]:           steps/mtaBuild.md
[resources-step-cloudFoundryDeploy]: steps/cloudFoundryDeploy.md
[resources-scenarios]:               scenarios

[SAP Cloud Platform]:                [https://account.hana.ondemand.com]
[SAP Cloud Platform Trial]:          [https://account.hanatrial.ondemand.com]
[devops-docker-images]:              https://github.com/SAP/devops-docker-images
[devops-docker-images-issues]:       https://github.com/SAP/devops-docker-images/issues
[cloud-cf-helloworld-nodejs]:        https://github.com/SAP/cloud-cf-helloworld-nodejs
[sap]:                               https://www.sap.com
[github]:                            https://github.com
[jenkins-io-documentation]:          https://jenkins.io/doc/
[jenkins-io-jenkinsfile]:            https://jenkins.io/doc/book/pipeline/jenkinsfile

