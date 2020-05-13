# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* This step is for triggering an ATC run on an ABAP system.
* You can either provide the ABAP endpoint config to directly trigger ann ATC run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP Cloud Platform ABAP Environment instance in Cloud Foundry that contains all the details to trigger an ATC run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key you have to provide the configuration of the packages and software components you want to be checked in an ATC run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile defining the pipeline.

Examples will be listed below.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

* ### ATC run via Cloud Foundry Service Key example in Jenkinsfile

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
    atcConfig: 'atcconfig.yml',
    script: this,
)
```

To trigger the ATC run an ATC config file `atcconfig.yml` will be needed. Check section 'ATC config file example' for more information.

* ### ATC run via direct ABAP endpoint configuration in Jenkinsfile

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

* ### ATC config file example

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
