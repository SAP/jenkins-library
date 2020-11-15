# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

- On SAP Cloud Platform, Cloud Foundry needs to be enabled on subaccount level. This can be done on the Subaccount Overview page. The subaccount is then mapped to a “Cloud Foundry Organization”, for which you must provide a suitable name during the creation. Have a look at the [documentation](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/dc18bac42270468d84b6c030a668e003.html) for more details.
- A (technical) user is required to access the SAP Cloud Platform via the cf CLI. The user needs to be a [member of the global account](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/4a0491330a164f5a873fa630c7f45f06.html) and has to have the [Space Developer](https://help.sap.com/viewer/a96b1df8525f41f79484717368e30626/Cloud/en-US/967fc4e2b1314cf7afc7d7043b53e566.html) role. The user and password need to be stored in the Jenkins Credentials Store.
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
abapEnvironmentCloneGitRepo (
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

## Example: Configuration of an abap-oem Service

If you want to use an abap-oem System please note that the following configurations need to be specified correctly:

* `abapSystemParentSaasAppname`: needs to fulfill the syntax [a-zA-Z0-9\-\_]+
* `abapSystemConsumerTenantLimit`: The number of consumer tenants tmust be greater or equal 1
* Either `abapSystemParentServiceLabel` or `abapSystemParentSaasAppname` must be set - depending on who created the oem-instance

You can either specify the parameters required for an abap-oem System either directly in the Jenkinsfile or in a `manifest.yml` file.

The first example makes use of the Jenkinsfile. 

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
    abapSystemParentServiceLabel: "abap-trial"
    abapSystemParentServiceInstanceGuid: "131bb94b-3045-4303-94bc-34df92072302"
    abapSystemParentSaasAppname: "abapcp-saas-itapcao1"
    abapSystemParentServiceParameters: `{"foo":"bar","veryspecialfeature":"true"}`
    abapSystemConsumerTenantLimit: 1

```

In this second example a configuration file is used that can be passed in the Jenkinsfile. 

```groovy
abapEnvironmentCloneGitRepo (
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
  parameters: "{ \"admin_email\" : \"user@example.com\", \"description\" : \"ABAP Environment Q System\", \"is_development_allowed\" : true, \"sapsystemname\" : \"H02\", \"size_of_persistence\" : 4, \"size_of_runtime\" : 1,
  \"parent_service_label\":\"abap-trial\",\"parent_service_instance_guid\":\"131bb94b-3045-4303-94bc-34df92072302\",\"parent_saas_appname\":\"abapcp-saas-itapcao1\",\"parent_service_parameters\":\"{\"foo\":\"bar\",\"veryspecialfeature\":\"true\"}\",\"consumer_tenant_limit\":1 }"
```
