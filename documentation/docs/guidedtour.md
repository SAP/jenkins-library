# Getting Started

##Prerequisites

* Linux - A Linux System with at least 4GB memory. All our samples were tested on Ubuntu 16.04. On Microsoft Windows you might face issues.
* Docker - All tests were made on Docker 18.09.6. Install the newest version from [docker.com](https://docs.docker.com/install/)
* Jenkins - Jenkins version 2.60.3 or higher. We recommend to use the CX-Server toolkit.  
* Access to [github.com][github].


##Jenkins

We offer the life-cycle management toolkit around Jenkins named `cx-server` to ease its usage and configuration. Based on docker images, you will get a preconfigured Jenkins and a Nexus based cache. 
Optionally, you can still use your [own Jenkins installation](#my-own-jenkins).

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

###My own Jenkins

tbd 

# Creating your first Pipeline
Open the Jenkins UI and Click the New Item menu. 

![Clicke New Item](../images/JenkinsHomeMenu-1.png "Jenkins Home Menu")

Provide a name for your new item (e.g. My Pipeline) and select `Pipeline`

![Create Pipeline Job](../images/JenkinsNewItemPipeline-1.png "Jenkins New Item")

Click the Add Source button, choose the type of repository you want to use and fill in the details.

Click the Save button and watch your first Pipeline run!

# Pipeline Configuration

# Running your Pipeline


text

[github]: https://github.com
[devops-docker-images-issues]: https://github.com/SAP/devops-docker-images/issues
[license]: LICENSE
[contribution]: CONTRIBUTING.md

