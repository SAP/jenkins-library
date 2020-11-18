# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* A SAP Cloud Platform ABAP Environment system is available. On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario “SAP Cloud Platform ABAP Environment - Software Component Test Integration (SAP_COM_0510)“. This can be done manually through the respective applications on the SAP Cloud Platform ABAP Environment System or through creating a service key for the system on cloud foundry with the parameters {“scenario_id”: “SAP_COM_0510", “type”: “basic”}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You can either provide the ABAP endpoint configuration to directly trigger an ATC run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP Cloud Platform ABAP Environment system in Cloud Foundry that contains all the details of the ABAP endpoint to trigger an ATC run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key you have to provide the configuration of the packages and software components you want to be checked in an ATC run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile defining the pipeline.
* The Software Components and/or Packages you want to be checked must be present in the configured system in order to run the check. Please make sure that you have created or pulled the respective Software Components and/or Packages in the SAP Cloud Platform ABAP Environment system.

Examples will be listed below.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentRunATCCheck script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentRunATCCheck:
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    atcConfig: 'atcconfig.yml',
```

### ATC run via Cloud Foundry Service Key example in Jenkinsfile

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
    cfServiceKeyName: 'myServiceKey',
    abapCredentialsId: 'cfCredentialsId',
    atcConfig: 'atcconfig.yml',
    script: this,
)
```

To trigger the ATC run an ATC config file `atcconfig.yml` will be needed. Check section 'ATC config file example' for more information.

### ATC run via direct ABAP endpoint configuration in Jenkinsfile

This  example triggers an ATC run directly on the ABAP endpoint.

In order to trigger the ATC run you have to pass the username and password for authentication to the ABAP endpoint via parameters as well as the ABAP endpoint/host. You can store the credentials in Jenkins and use the abapCredentialsId parameter to authenticate to the ABAP endpoint/host.

This must be configured as following:

```groovy
abapEnvironmentRunATCCheck(
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    atcConfig: 'atcconfig.yml',
    script: this,
)
```

To trigger the ATC run an ATC config file `atcconfig.yml` will be needed. Check section 'ATC config file example' for more information.

### ATC config file example

The following section contains an example of an `atcconfig.yml` file.
This file must be stored in the same Git folder where the `Jenkinsfile` is stored to run the pipeline. This folder must be taken as a SCM in the Jenkins pipeline to run the pipeline.

You can specify a list of packages and/or software components to be checked. This must be in the same format as below example for a `atcconfig.yml` file.
For each package that has to be checked you can configure if you want the subpackages to be included in checks or not.
Please note that if you chose to provide both packages and software components to be checked with the `atcconfig.yml` file, the set of packages and the set of software components will be combinend by the API using a logical AND operation.
Therefore, we advise to specify either the Software Components or Packages.

See below example for an `atcconfig.yml` file with both packages and software components to be checked:

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

The following example of an `atcconfig.yml` file that only contains packages to be checked:

```yaml
atcobjects:
  package:
    - name: "TestPackage"
      includesubpackage: false
    - name: "TestPackage2"
      includesubpackage: true
```

The following example of an `atcconfig.yml` file that only contains software components to be checked:

```yaml
atcobjects:
  softwarecomponent:
    - name: "TestComponent"
    - name: "TestComponent2"
```

The following is an example of an `atcconfig.yml` file that supports the check variant and configuration ATC options:

```yaml
checkvariant: "TestCheckVariant"
configuration: "TestConfiguration"
atcobjects:
  softwarecomponent:
    - name: "TestComponent"
    - name: "TestComponent2"
```
