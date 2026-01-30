# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites SAP BTP, ABAP environment

* A SAP BTP, ABAP environment system is available.
  * This can be created manually on Cloud Foundry.
  * In a pipeline, you can do this, for example, with the step [cloudFoundryCreateService](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateService/).
* Communication Scenario [“SAP BTP, ABAP Environment - Software Assembly Integration (SAP_COM_0582)“](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/26b8df5435c649aa8ea7b3688ad5bb0a.html) is setup for this system.
  * E.g. a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) are configured.
  * This can be done manually through the respective applications on the SAP BTP, ABAP environment system,
  * or through creating a service key for the system on cloud foundry with the parameters {“scenario_id”: “SAP_COM_0582", “type”: “basic”}.
  * In a pipeline, you can do this, for example, with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).
* You have following options to provide the ABAP endpoint configuration:
  * The host and credentials the SAP BTP, ABAP environment system itself. The credentials must be configured for the Communication Scenario SAP_COM_0582.
  * The Cloud Foundry parameters (API endpoint, organization, space), credentials, the service instance for the ABAP service and the service key for the Communication Scenario SAP_COM_0582.
  * Only provide one of those options with the respective credentials. If all values are provided, the direct communication (via host) has priority.

## Prerequisites On Premise

* You need to specify the host and credentials to your system
* A certificate for the system needs to be stored in .pipeline/trustStore and the name of the certificate needs to be handed over via the configuration

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

If you want to use this step several time in one pipeline with different phases, the steps have to be put in different stages as it is not allowed to run the same step repeatedly in one stage.

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
stage('MyPhase') {
            steps {
                abapEnvironmentBuild script: this
            }
        }
```

If you want to provide the host and credentials of the Communication Arrangement directly or you want to run in on premise, the configuration could look as follows:

```yaml
stages:
  MyPhase:
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
```

Or by authenticating against Cloud Foundry and reading the Service Key details from there:

```yaml
stages:
  MyPhase:
    abapCredentialsId: 'cfCredentialsId',
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
    cfServiceKeyName: 'myServiceKey',
```

One possible complete config example. Please note that the values are handed over as a string, which has inside a json structure:

```yaml
stages:
  MyPhase:
    abapCredentialsId: 'abapCredentialsId'
    host: 'https://myABAPendpoint.com'
    certificateNames: ['myCert.cer']
    phase: 'MyPhase'
    values: '[{"value_id":"ID1","value":"Value1"},{"value_id":"ID2","value":"Value2"}]'
    downloadResultFilenames: ['File1','File2']
    publishResultFilenames: ['File2']
    subDirectoryForDownload: 'MyDir'
    filenamePrefixForDownload: 'MyPrefix'
    treatWarningsAsError: true
    maxRuntimeInMinutes: 360
    pollingIntervallInSeconds: 15
```
