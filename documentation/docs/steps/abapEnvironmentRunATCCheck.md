# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* This step is for triggering an ATC run on an ABAP system.
* You can either provide the ABAP endpoint config to directly trigger ann ATC run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of an ABAP instance in Cloud Foundry that contains all the details to trigger an ATC run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key you have to provide the configuration of the packages and software components you want to be checked in an ATC run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile where you run the pipeline from.
Examples will be listed below.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

The following example triggers an ATC run via reading the Service Key of an ABAP instance in Cloud Foundry.
You can store the credentials in Jenkins and use the cfCredentialsId parameter to authenticate to Cloud Foundry.
The username and password to authenticate to ABAP system will then be read from the Cloud Foundry Service Key that is bound to the ABAP instance.
This can be done accordingly:

```groovy
    abapEnvironmentRunATCCheck(
        cfApiEndpoint : 'https://test.server.com',
        cfOrg : 'cfOrg',
        cfSpace: 'cfSpace',
        cfServiceInstance: 'myServiceInstance',
        cfSserviceKeyName: 'myServiceKey',
        cfCredentialsId: 'cfCredentialsId',
		atcrunConfig: 'atcrunconfig.yml',
        script: this,
    ) 
```

This  example triggers an ATC run directly on the ABAP endpoint.
In order to trigger the ATC run you have to pass the username and password for authentication to the ABAP endpoint via parameters.
That must be configured as following:

```groovy
    abapEnvironmentRunATCCheck(
		username: 'myUser',
		password: 'myPassword',
		host: 'https://myABAPendpoint.com',
        cfCredentialsId: 'cfCredentialsId',
		atcrunConfig: 'atcrunconfig.yml',
        script: this,
    ) 
```

The following section contains an example of an atcrunconfig.yml file.
You can specify a list of packages and software components to be checked. This must be in the same format as below.
For each package that has to be checked you can configure if you want the subpackages to be included in checks or not.
See below example:

```yaml
atcobjects:
  package:
    - name: "TestPackage"
      includesubpackage: false
    - name: "TestPackage2"
      includesubpackage: true
  softwarecomponent:
    - name: "TestComponent"
    - name: "TestComponent2"
```