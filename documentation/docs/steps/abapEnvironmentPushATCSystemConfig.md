# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* A SAP BTP, ABAP environment system is available. On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario “ABAP Test Cockpit Configuration Integration (SAP_COM_0763)“. This can be done manually through the respective applications on the SAP BTP, ABAP environment system or through creating a service key for the system on Cloud Foundry with the parameters {“scenario_id”: “SAP_COM_0763", “type”: “basic”}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You can either provide the ABAP endpoint configuration to directly trigger an ATC run on the ABAP system or optionally provide the Cloud Foundry parameters with your credentials to read a Service Key of a SAP BTP, ABAP environment system in Cloud Foundry that contains all the details of the ABAP endpoint to trigger an ATC run.
* Regardless if you chose an ABAP endpoint directly or reading a Cloud Foundry Service Key, you have to provide the configuration of the packages and software components you want to be checked in an ATC run in a .yml or .yaml file. This file must be stored in the same folder as the Jenkinsfile defining the pipeline.
* The software components and/or packages you want to be checked must be present in the configured system in order to run the check. Please make sure that you have created or pulled the respective software components and/or Packages in the SAP BTP, ABAP environment system.

Examples will be listed below.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentPushATCSystemConfig script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentPushATCSystemConfig:
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    atcSystemConfigFilePath: 'atcSystemConfig.json',
```

To trigger a create/update ATC System Configuration step an ATC System configuration file `atcSystemConfig.json` will be needed. Check section 'ATC System Configuration file example' for more information.

### Create/Update an ATC System Configuration via Cloud Foundry Service Key example in Jenkinsfile

The following example triggers a Create/Update operation on an ATC System Configuration via reading the Service Key of an ABAP instance in Cloud Foundry.

You can store the credentials in Jenkins and use the cfCredentialsId parameter to authenticate to Cloud Foundry.
The username and password to authenticate to ABAP system will then be read from the Cloud Foundry service key that is bound to the ABAP instance.

This can be done accordingly:

```groovy
abapEnvironmentPushATCSystemConfig(
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
    cfServiceKeyName: 'myServiceKey',
    abapCredentialsId: 'cfCredentialsId',
    atcSystemConfigFilePath: 'atcSystemConfig.json',
    script: this,
)
```

### Create/Update an ATC System Configuration via direct ABAP endpoint configuration in Jenkinsfile

This example triggers a create/update operation on an ATC System Configuration run directly on the ABAP endpoint.

In order to trigger the create/update operation on an ATC System Configuration you have to pass the username and password for authentication to the ABAP endpoint via parameters as well as the ABAP endpoint/host. You can store the credentials in Jenkins and use the abapCredentialsId parameter to authenticate to the ABAP endpoint/host.

This must be configured as following:

```groovy
abapEnvironmentPushATCSystemConfig(
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
    atcSystemConfigFilePath: 'atcSystemConfig.json',
    script: this,
)
```

To create/update an ATC System Configuration a file `atcSystemConfig.json` will be needed. Check section 'ATC System configuration file example' for more information.

### ATC System configuration file example

The step always performs a check first, if an ATC System Configuration with the same name provided in the file `atcSystemConfig.json` with the attribute conf_name.
This file contains an JSON Representation of an ATC System Configuration. Some json file examples can be found below.

In case an ATC System Configuration with this name already exists, by default, the step would perform an update of this ATC System Configuration with the ATC System Configuration information provided in file `atcSystemConfig.json`.
If this is not desired, an update could be supressed by using the parameter patchIfExisting in the configuration yaml the following way:

```yaml
steps:
  abapEnvironmentPushATCSystemConfig:
    atcSystemConfigFilePath: atcSystemConfig.json,
    patchIfExisting: false,
```

In this case the step skips further processing after existence check and returns with a Warning.

Providing a specifc System configuration file `atcSystemConfig.json` is mandatory.

The following section contains an example of an `atcSystemConfig.json` file.

This file must be stored in the same Git folder where the `Jenkinsfile` is stored to run the pipeline. This folder must be taken as a SCM in the Jenkins pipeline to run the pipeline.

See below an example for an `atcSystemConfig.json` file for creating/updating an ATC System Configuration with the name myATCSystemConfigurationName including a change of one priority.

```json
{
  "conf_name": "myATCSystemConfigurationName",
  "checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
  "block_findings": "0",
  "inform_findings": "1",
  "_priorities": [
    {
      "test": "CL_CI_TEST_AMDP_HDB_MIGRATION",
      "message_id": "FAIL_ABAP",
      "priority": 2
    }
  ]
}
```

See below an example for an `atcSystemConfig.json` file for creating/updating an ATC System Configuration with the name myATCSystemConfigurationName.

```json
{
  "conf_name": "myATCSystemConfigurationName",
  "checkvariant": "SAP_CLOUD_PLATFORM_ATC_DEFAULT",
  "block_findings": "0",
  "inform_findings": "1"
}
```
