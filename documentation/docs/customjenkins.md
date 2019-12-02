# Custom Jenkins Setup

Although we recommend to use the Cx server, you don't need to. You can run project "Piper" on your own Jenkins installation. However, you have to care for some settings the Cx Server gives you for free. Furthermore, the support of a none Cx Server installations is challenging. 	

This section describes the necessary adjustments you might have to take.

## Requirements

* Java Runtime Environment 8
* Installation of Jenkins v 2.60.3 or higher running on Linux. We tested with debian-stretch.
* A Jenkins user with administration privileges.
* The Jenkins instance has access to [github.com][github].

## Docker
Most tools used by project "Piper" to build, test and deploy your application are available as out-of-the box Docker images. No need to manually install them neither on your Jenkins server nor on your Jenkins nodes, and no need to care for the updates. Instead these are pulled from *hub.docker.com* on usage.

Install Docker if you haven't yet. See [Docker Community Edition][docker-install] to install the newest version of Docker.
**Note:** we have tested on Docker 18.09.6.

If your Jenkins server already runs as Docker container make sure the tools container can run on the Docker host. Extend the Docker call in the following way:

```
docker run ...  -v /var/run/docker.sock:/var/run/docker.sock ...
```

## Plugins 

Project "Piper" requires a set of plugins installed on your Jenkins server. This set may evolve in the future. Make sure that all plugins of the appropriate versions are installed.

The Cx server repository contains an [up-to-date list][devops-cxs-plugins] of the required plugins. You could download the list 

```
curl -o plugins.txt https://raw.githubusercontent.com/SAP/devops-docker-cx-server/master/jenkins-master/plugins.txt
```

and use the [Jenkins client][jenkins-doc-client] to ease the installation. Run the following command on the Jenkins server with an user holding administrative privileges:

```
cat plugins.txt | awk '{system("java " "-jar jenkins-cli.jar -s http://localhost:8080 -auth ${ADM_USER}:${ADM_PASSWD} install-plugin " $1)}'
```

## Shared Library

Shared libraries extending the Jenkins pipeline are defined within the Jenkins system configuration. A library is defined by a link to its source repository and an appropriate version identifier. To add the project "Piper"s library:

1. Open the Jenkins UI under `http://<jenkins-server-address>:<http-port>`, login with administration privileges and choose **Manage Jenkins > Configure System**.

   <p align="center">
   ![Configure System](images/JenkinsHomeMenuManageConfig.png "Configure System")
   </p>

1. Scroll down to section **Global Pipeline Libraries** and choose the **Add** button. A new library is created.

   <p align="center">
   ![Add Library](images/JenkinsConfigSystemLibrary-Add.png "Add Library")
   </p>

1. For **Library Name** enter the library name `piper-lib-os`

1. For **Default Version** enter the branch or tag you want to consume (e.g. `master` or `v0.1`)

1. For **Retrieval Method** choose **Modern SCM**

1. For **Source Code Management** choose **Git**

1. For **Project Repository** enter the GitHub URL of the project Piper shared library `https://github.com/SAP/jenkins-library`

   <p align="center">
   ![Library Setup](images/JenkinsConfigSystemLibrary-Edit.png "Library Setup")
   </p>

1. Save changes

**Result:** The library is available as `piper-lib-os` and can be used in any pipeline by adding the following line to its `Jenkinsfile`:

```groovy
@Library('piper-lib-os') _
```

When the pipeline is launched, Jenkins downloads the corresponding library as source and compiles it before the pipeline is processed.


## User Permission Issue

Your native Jenkins installation defines the user `jenkins` as service user. If the user doesn't exists it will be created. The user id will be the next free number determined by `/etc/passwd` - probably starting from `100`.
In contrast, the official [Jenkins Docker image][jenkins-docker-image] defines the user `jenkins` with the userid `1000` as service user inside the container. 
So, the service user id of your native Jenkins server most likely differs from the user id of the official Jenkins Docker image.

This could have impacts. 

Project "Piper" is running many pipeline steps as Docker images. If a Docker container is created, the Jenkins Docker plugin passes the Jenkins user and group id as process owner into the Docker container.
Binding a folder from the host machine into the container - used to exchange files between steps - results in file permission issues if the user used inside the container doesn't have rights to the folder on the host machine or vice versa.

You wont face this issue with images of the project "Piper" but, some 3rd party docker images follow this convention and expect to be executed under userid `1000`, like [node.js][dockerhub-node] which is used by a set of additional steps. 

If you run into such [user permission issue][piper-issue-781] you have following options

1. Use id `1000` - change the id of your Jenkins service user to `1000`.

1. [Create your own images][docker-getstarted] - derive from the Docker image in question and solve the permission issues by removing the file system restrictions. Adjust the configuration accordingly, e.g.adjust the `npmExecute` step of your project's YAML:

   ```
     npmExecute:
       dockerImage: 'my-node:8-stretch'
   ```

1. Setup a namespace - the user permission [issue 781][piper-issue-781] of the piper repository described how to setup a Linux kernel user namespace to circumvent the mismatch of user ids. This solution is experimental and should be well-considered.


[github]: https://github.com
[docker-install]: https://docs.docker.com/install
[dockerhub-node]: https://hub.docker.com/_/node/
[docker-getstarted]: https://docs.docker.com/get-started/
[jenkins-doc-client]: https://jenkins.io/doc/book/managing/cli/
[jenkins-docker-image]: https://github.com/jenkinsci/docker/
[piper-library-pages]: https://sap.github.io/jenkins-library
[piper-issue-781]: https://github.com/SAP/jenkins-library/issues/781

[devops-docker-images]: https://github.com/SAP/devops-docker-images
[devops-cxs-plugins]: https://github.com/SAP/devops-docker-cx-server/blob/master/jenkins-master/plugins.txt


