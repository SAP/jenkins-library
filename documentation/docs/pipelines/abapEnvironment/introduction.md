# ABAP Environment Pipeline

The goal of the ABAP Environment Pipeline is to enable Continuous Integration for the SAP Cloud Platform ABAP Environment, also known as Steampunk.
In the current state, the pipeline enables you to pull Software Components to specifc systems and perform ATC checks. The following steps are performed:

* Create an instance of the SAP Cloud Platform ABAP Environment service
* Configure the Communication Arrangement SAP_COM_0510
* Pull Git repositories / Software Components to the instance
* Run ATC Checks
* Delete the SAP Cloud ABAP Environment system

## Configuration

### 1. Jenkins Server

Configure your Jenkins Server according to the [documentation](https://sap.github.io/jenkins-library/guidedtour/)

### 2. Jenkinsfile

Create a file named `Jenkinsfile` in your repository with the following content:

```
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

The annotation `@Library('piper-lib-os')` is a reference to the Jenkins Configuration, where you configured the Piper Library as a "Global Pipeline Library". If you want to **avoid breaking changes** we advise you to use a specific release of the Piper Library instead of the default master branch (see [documentation](https://sap.github.io/jenkins-library/customjenkins/#shared-library))

### 3. Manifest for Service Creation

Create a file `manifest.yml`. The pipeline will create a SAP Cloud Platform ABAP Environment System in the beginning (and delete it in the end). This file describes the ABAP instance, which will be created:

```yaml
---
create-services:
- name:   "abapEnvironmentPipeline"
  broker: "abap"
  plan:   "16_abap_64_db"
  parameters: "{ \"admin_email\" : \"user@example.com\", \"description\" : \"System for ABAP Pipeline\" }"
```

The example values are a suggestion. Please change them accordingly and don't forget to enter your own email address. Please be aware that creating a SAP Cloud ABAP Environment instance may incur costs.

### 4. Configuration for the Communication

The communication to the ABAP system is done using a Communication Arrangement. The Communication Arrangement is created during the pipeline via the command `cf create-service-key`. The configuration for the command needs to be stored in a JSON file. Create the file `sap_com_0510.json` in the repository with the following content:

```json
{
  "scenario_id": "SAP_COM_0510",
  "type": "basic"
}
```

### 5. Configuration for ATC

Create a file `atcConfig.yml` to store the configuration for the ATC run. In this file, you can specify which Packages or Software Components shall be checked. Please have a look at the step documentation for more details. Here is an example of the configuration:

```yml
atcobjects:
  softwarecomponent:
    - name: "/DMO/REPO"
```

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/) for more details

### 6. Technical Pipeline Configuration

Create a file `.pipeline/config.yml` where you store the configuration for the pipeline, e.g. apiEndpoints and credentialIds. The steps make use of the Credentials Store of the Jenkins Server. Here is an example of the configuration file:

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
steps:
  cloudFoundryDeleteService:
    deleteServiceKeys: true
```

If one stage of the pipeline is not configured in this yml file, the stage will not be executed during the pipeline run. If the stage `Prepare System` is configured, the system will be deprovisioned in the cleanup routine - although it is necessary to configure the steps `cloudFoundryDeleteService` as above.

## Extension

You can extend each stage of this pipeline following the [documentation](../../extensibility.md).

For example, this can be used to display ATC results utilizing the checkstyle format with the [Warnings Next Generation Plugin](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#warnings-next-generation-plugin) ([GitHub Project](https://github.com/jenkinsci/warnings-ng-plugin)).
To achieve this, create a file `.pipeline/extensions/ATC.groovy` with the following content:

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  recordIssues tools: [checkStyle(pattern: '**/ATCResults.xml')], qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

While `tools: [checkStyle(pattern: '**/ATCResults.xml')]` will display the ATC findings using the checkstyle format, `qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]` will set the build result to UNSTABLE in case the ATC results contain at least one warning or error.
