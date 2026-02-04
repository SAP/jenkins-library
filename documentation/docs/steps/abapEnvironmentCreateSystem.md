# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

- On SAP Business Technology Platform (SAP BTP), Cloud Foundry needs to be enabled on subaccount level. This can be done on the Subaccount Overview page. The subaccount is then mapped to a “Cloud Foundry Organization”, for which you must provide a suitable name during the creation. Have a look at the [documentation](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/dc18bac42270468d84b6c030a668e003.html) for more details.
- A (technical) user is required to access the SAP BTP via the cf CLI. The user needs to be a [member of the global account](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/4a0491330a164f5a873fa630c7f45f06.html) and has to have the [Space Developer](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/967fc4e2b1314cf7afc7d7043b53e566.html) role. The user and password need to be stored in the Jenkins Credentials Store.
- Please make sure, that there are enough entitlements in the subaccount for the [Service Plan](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/c40cb18aeaa343389036fdcdd03c41d0.html), which you want to use for this step.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example: Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentCreateSystem script: this
```

The configuration values for the system can be passed through the `config.yml` file:

```yaml
steps:
  abapEnvironmentCreateSystem:
    cfCredentialsId: 'cfCredentialsId'
    cfApiEndpoint: 'https://test.server.com'
    cfOrg: 'cfOrg'
    cfSpace: 'cfSpace'
    cfServiceInstance: 'H02_Q_system'
    cfService: 'abap'
    cfServicePlan: 'standard'
    abapSystemAdminEmail: 'user@example.com'
    abapSystemDescription: 'ABAP Environment Q System'
    abapSystemIsDevelopmentAllowed: true
    abapSystemID: 'H02'
    abapSystemSizeOfPersistence: 4
    abapSystemSizeOfRuntime: 1
```

## Example: Configuration in the Jenkinsfile

The step, including all parameters, can also be called directly from the Jenkinsfile. In the following example, a configuration file is used.

```groovy
abapEnvironmentCreateSystem (
  script: this,
  cfCredentialsId: 'cfCredentialsId',
  cfApiEndpoint: 'https://test.server.com',
  cfOrg: 'cfOrg',
  cfSpace: 'cfSpace',
  cfServiceManifest: 'manifest.yml'
)
```

The file `manifest.yml` would look like this:

```yaml
---
create-services:
- name:   "H02_Q_system"
  broker: "abap"
  plan:   "standard"
  parameters: "{ \"admin_email\" : \"user@example.com\", \"description\" : \"ABAP Environment Q System\", \"is_development_allowed\" : true, \"sapsystemname\" : \"H02\", \"size_of_persistence\" : 4, \"size_of_runtime\" : 1 }"
```
