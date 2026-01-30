# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

A SAP BTP, ABAP environment system is available.
On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario "Software Component Management Integration (SAP_COM_0948)". This can be done manually through the respective applications on the SAP BTP, ABAP environment system or through creating a service key for the system on Cloud Foundry with the parameters {"scenario_id": "SAP_COM_0948", "type": "basic"}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/). In addition, the software component should be cloned into the system instance. You can do this with the step [abapEnvironmentCloneGitRepo](./abapEnvironmentCloneGitRepo.md).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example: Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentCreateTag script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentCreateTag:
    repositoryName: '/DMO/GIT_REPOSITORY'
    commitID: 'cd87a3c'
    tagName: 'myTag'
    tagDescription: 'Created via Jenkins'
    abapCredentialsId: 'abapCredentialsId'
    host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
```

Another option is to read the host and credentials from the cloud foundry service key of the respective instance. Furthermore, if you want to specify multiple repositories, you can use a configuration file: `repositories.yml`/`addon.yml`. If you are using the ABAP Environment Pipeline to [build an add-on](../scenarios/abapEnvironmentAddons.md), you can also generate tags based on the product and component versions.

With this approach the `config.yml` would look like this:

```yaml
steps:
  abapEnvironmentCreateTag:
    repositories: 'repositories.yml'
    generateTagForAddonProductVersion: true
    generateTagForAddonComponentVersion: true
    cfCredentialsId: 'cfCredentialsId'
    cfApiEndpoint: 'https://test.server.com'
    cfOrg: 'cfOrg'
    cfSpace: 'cfSpace'
    cfServiceInstance: 'cfServiceInstance'
    cfServiceKeyName: 'cfServiceKeyName'
```

and the configuration file `repositories.yml`/`addon.yml` would look like this:

```yaml
addonVersion: "1.2.3"
addonProduct: "/DMO/PRODUCT"
repositories:
  - name: '/DMO/REPO'
    branch: 'feature'
    commitID: 'cd87a3cac2bc946b7629580e58598c3db56a26f8'
    version: '1.0.0'
```

Using such a configuration file is the recommended approach. Please note that you need to use the YAML data structure as in the example above when using the `repositories.yml`/`addon.yml` config file.
