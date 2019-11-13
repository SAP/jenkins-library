# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* A SAP Cloud Platform ABAP Environment system is available.
* On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario "SAP Cloud Platform ABAP Environment - Software Component Test Integration (SAP_COM_0510)".

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
abapEnvironmentPullGitRepo (
    host : 'https://1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY',
    repositoryName : '/DMO/GIT_REPOSITORY',
    credentialsId : "myCredentialsId",
    script : this
)
```
