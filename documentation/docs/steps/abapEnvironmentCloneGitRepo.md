# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

A SAP Cloud Platform ABAP Environment system is available.
On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario "SAP Cloud Platform ABAP Environment - Software Component Test Integration (SAP_COM_0510)". This can be done manually through the respective applications on the SAP Cloud Platform ABAP Environment System or through creating a service key for the system on cloud foundry with the parameters {"scenario_id": "SAP_COM_0510", "type": "basic"}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/). In addition, the software component should be cloned into the system instance. You can do this with the step [abapEnvironmentPullGitRepo](https://sap.github.io/jenkins-library/steps/abapEnvironmentPullGitRepo/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example: Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentCheckoutBranch script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml

steps:
  abapEnvironmentCloneGitRepo:
    repositoryName: '/DMO/GIT_REPOSITORY'
    branchName: 'my-demo-branch'
    abapCredentialsId: 'abapCredentialsId'
    host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
```
Please note that the branchName parameter specifies the target branch you want to clone. Also keep in mind that the repositoryName parameter must define a single repository.

Another option is to read the host and credentials from the cloud foundry service key of the respective instance. Furthermore, if you want to clone multiple repositories, they can be specified in a configuration file.

With this approach the `config.yml` would look like this:

```yaml
steps:
  abapEnvironmentCloneGitRepo:
    repositories: 'repositories.yml'
    cfCredentialsId: 'cfCredentialsId'
    cfApiEndpoint: 'https://test.server.com'
    cfOrg: 'cfOrg'
    cfSpace: 'cfSpace'
    cfServiceInstance: 'cfServiceInstance'
    cfServiceKeyName: 'cfServiceKeyName'
```

and the configuration file `repositories.yml` would look like this:

```yaml
repositories:
  - name: '/DMO/GIT_REPOSITORY'
    branch: 'master'
  - name: '/DMO/SOFTWARE_COMPONENT'
    branch: 'feature'
```

Using such a configuration file is the recommended approach. Please note that you need to use the YAML data structure as in the example above when using the `repositories.yml` config file.

!!! note "Commit IDs"
    CommitIDs will also be supported in the future. While the step already includes the handling of commit IDs, the ABAP Environment system will support this not until a later release.

## Example: Configuration in the Jenkinsfile

It is also possible to call the steps - including all parameters - directly in the Jenkinsfile.
In the first example, the host and the credentialsId of the Communication Arrangement are directly provided.

```groovy
abapEnvironmentCloneGitRepo (
  script: this,
  repositoryName: '/DMO/GIT_REPOSITORY',
  branchName: 'my-demo-branch',
  abapCredentialsId: 'abapCredentialsId',
  host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
)
```

In the second example, the host and credentialsId will be read from the provided cloud foundry service key of the specified service instance.

```groovy
abapEnvironmentCloneGitRepo (
  script: this,
  repositoryName: '/DMO/GIT_REPOSITORY',
  branchName: 'my-demo-branch'
  abapCredentialsId: 'cfCredentialsId',
  cfApiEndpoint: 'https://test.server.com',
  cfOrg: 'cfOrg',
  cfSpace: 'cfSpace',
  cfServiceInstance: 'cfServiceInstance',
  cfServiceKeyName: 'cfServiceKeyName'
)
```
