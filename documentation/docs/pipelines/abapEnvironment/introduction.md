# ABAP Environment Pipeline

The goal of the ABAP Environment Pipeline is to enable Continuous Integration for the SAP Cloud Platform ABAP Environment, also known as Steampunk.
In the current state, the pipeline enables you to pull your Software Components to specifc systems and perform ATC checks. The following steps are performed:
 * Create an instance of the SAP Cloud Platform ABAP Environment
 * Configure the Communication Arrangement SAP_COM_0510
 * Pull Git repositories / Software Components to the instance
 * Run ATC Checks
 * Delete the SAP Cloud ABAP Environment system

## Configuration

1. Configure your Jenkins Server according to the [documentation](https://sap.github.io/jenkins-library/guidedtour/)
2. Create a file named `Jenkinsfile` in your repository with the following content:

```
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

The annotation `@Library('piper-lib-os')` is a reference to the Jenkins Configuration, where you configured the Piper Library as a "Global Pipeline Library". If you want to **avoid breaking changes** we advise you to use a specific release of the Piper Library instead of the default master branch (see https://sap.github.io/jenkins-library/customjenkins/#shared-library)

3. Create a file `manifest.yml`. The pipeline will create a SAP Cloud Platform ABAP Environment System in the beginning (and delete it in the end). This file describes the ABAP instance, which will be created:
```yml
---
create-services:
- name:   "abapEnvironmentPipeline"
  broker: "abap"
  plan:   "16_abap_64_db"
  parameters: "{ \"admin_email\" : \"user@example.com\", \"description\" : \"System for ABAP Pipeline\" }"
```
Please be aware that creating a SAP Cloud ABAP Environment instance may incur costs.

4. The communication to the ABAP system is done using a Communication Arrangement. The Communication Arrangement is created during the pipeline via the command `cf create-service-key`. The configuration for the command needs to be stored in a JSON file. Create the file `sap_com_0510.json` in the repository with the following content:
```json
{
  "scenario_id": "SAP_COM_0510",
  "type": "basic"
}
```

4. Create a file `atcConfig.yml` to store the configuration for the ATC run. In this file, you can specify which Packages or Software Components shall be checked. Please have a look at the step documentation for more details. Here is an example of the configuration:
```yml
atcobjects:
  softwarecomponent:
    - name: "/DMO/REPO"
```

6. Create a file `.pipeline/config.yml` where you store the configuration for the pipeline, e.g. apiEndpoints and credentialIds. The steps make use of the Credentials Store of the Jenkins Server. Here is an example of the configuration file:
```yml
general:
  cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com'
  cfOrg: 'your-cf-org'
  cfSpace: 'yourSpace'
  cfCredentialsId: 'cfAuthentification'
  cfServiceInstance: 'abapEnvironmentPipeline'
  cfServiceKeyName: 'jenkins_sap_com_0510'
stages:
  Prepare System:
    cfServiceManifest: 'manifest.yml'
    stashContent: ''
    cfServiceKeyConfig: 'sap_com_0510.json'
  Clone Repositories:
    repositoryNames: ['/DMO/REPO']
  ATC:
    atcConfig: 'atcConfig.yml'
```
If one stage of the pipeline is not configured in this yml file, the stage will not be executed during the pipeline run.

### Extension

A documentation on how to extend stages can be found [here](https://sap.github.io/jenkins-library/extensibility/). This can be used to display ATC results with the Checksytle Plugin.
