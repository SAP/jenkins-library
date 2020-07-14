# Configuration

In this section, you can learn how to create a configuration in a (GitHub) repository to run an ABAP Environment Pipeline.

## 1. Prerequisites

* Configure your Jenkins Server according to the [documentation](https://sap.github.io/jenkins-library/guidedtour/).
* Create a git repository on a host reachable by the Jenkinsserver (e.g. GitHub.com). The pipeline will be configured in this repository.
* A Cloud Foundry Organization & Space with the necessary entitlements are available
* A Cloud Foundry User & Password with the required authorizations in the Organization and Space are available. User and Password were saved in the Jenkins Credentials Store

## 2. Jenkinsfile

Create a file named `Jenkinsfile` in your repository with the following content:

```
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

The annotation `@Library('piper-lib-os')` is a reference to the Jenkins Configuration, where you configured the Piper Library as a "Global Pipeline Library". If you want to **avoid breaking changes** we advise you to use a specific release of the Piper Library instead of the default master branch. This can be achieved by either adapting the configuration (see [documentation](https://sap.github.io/jenkins-library/infrastructure/customjenkins/#shared-library)) or by specifying the release within the annotaion:

```
@Library('piper-lib-os@v1.53.0') _
```

An Overview of the releases of the Piper Library can be found [here](https://github.com/SAP/jenkins-library/releases).

## 3. Manifest for Service Creation

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

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateService/) for more details.

## 4. Configuration for the Communication

The communication to the ABAP system is done using a Communication Arrangement. The Communication Arrangement is created during the pipeline via the command `cf create-service-key`. The configuration for the command needs to be stored in a JSON file. Create the file `sap_com_0510.json` in the repository with the following content:

```json
{
  "scenario_id": "SAP_COM_0510",
  "type": "basic"
}
```

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/) for more details.

## 5. Configuration for ATC

Create a file `atcConfig.yml` to store the configuration for the ATC run. In this file, you can specify which Packages or Software Components shall be checked. Please have a look at the step documentation for more details. Here is an example of the configuration:

```yml
atcobjects:
  softwarecomponent:
    - name: "/DMO/REPO"
```

Please have a look at the [step documentation](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/) for more details.

## 6. Technical Pipeline Configuration

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
    cfServiceKeyConfig: 'sap_com_0510.json'
  Clone Repositories:
    repositoryNames: ['/DMO/REPO']
  ATC:
    atcConfig: 'atcConfig.yml'
steps:
  cloudFoundryDeleteService:
    deleteServiceKeys: true
```

If one stage of the pipeline is not configured in this yml file, the stage will not be executed during the pipeline run. If the stage `Prepare System` is configured, the system will be deprovisioned in the cleanup routine - although it is necessary to configure the step `cloudFoundryDeleteService` as above.

## 7. Create a Jenkins Pipeline

On your Jenkinsserver click on `New Item` to create a new pipeline. Provide a name and select the type `Pipeline`.
On the creation screen for the pipeline, scroll to the section `Pipeline` and select `Pipeline script from SCM`. Provide the URL (and credentials - if required) of the repository, in which you configured the pipeline. Make sure the `Script Path` points to your Jenkinsfile - if you created the Jenkinsfile according to the documentation above, the default value should be correct.

If you want to configure a build trigger, this can be done in the section of the same name. Here is one example: to run the pipeline every night, you can tick the box "Run periodically". In the visible input field, you can specify a shedule. Click on the questionsmark to read the documentation. The following example will result in the pipeline running every night between 3am and 4am.

```
H H(3-4) * * *
```

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

While `tools: [checkStyle(pattern: '**/**/ATCResults.xml')]` will display the ATC findings using the checkstyle format, `qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]` will set the build result to UNSTABLE in case the ATC results contain at least one warning or error.

### Stage Names

The stage name for the extension is usually the displayed name, e.g. `ATC.groovy` or `Prepare System.groovy`. One exception is the generated `Post` stage. While the displayed name is "Declarative: Post Actions", you can extend this stage using `Post.groovy`.
