# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

A SAP Cloud Platform ABAP Environment system is available.
On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario "SAP Cloud Platform ABAP Environment - Software Component Test Integration (SAP_COM_0510)". This can be done manually through the respective applications on the SAP Cloud Platform ABAP Environment System or through creating a service key for the system on cloud foundry with the parameters {"scenario_id": "SAP_COM_0510", "type": "basic"}.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

In the first example, the host and the credentialsId of the Communication Arrangement are directly provided.

```groovy
abapEnvironmentPullGitRepo (
  script: this,
  repositoryNames: ['/DMO/GIT_REPOSITORY'],
  credentialsId: 'abapCredentialsId',
  host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
)
```

In the second example, the host and credentialsId will be read from the provided cloud foundry service key of the specified service instance.

```groovy
abapEnvironmentPullGitRepo (
  script: this,
  repositoryNames: ['/DMO/GIT_REPOSITORY', '/DMO/GIT_REPO'],
  credentialsId: 'cfCredentialsId',
  cfApiEndpoint: 'https://test.server.com',
  cfOrg: 'cfOrg',
  cfSpace: 'cfSpace',
  cfServiceInstance: 'cfServiceInstance',
  cfServiceKey: 'cfServiceKey'
)
```
