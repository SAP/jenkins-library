# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have an SAP Solution Manager user account and the roles required for uploading. See [SAP Solution Manager Administration](https://help.sap.com/viewer/c413647f87a54db59d18cb074ce3dafd/7.2.12/en-US/11505ddff03c4d74976dae648743e10e.html).
* You have created a change document.
* You have installed the Change Management Client with the needed certificates. See [Change Management Client](transportRequestUploadSOLMAN.md#Change-Management-Client).

## Specifying the Change Document

The target of the status check is a change document, identified by an identifier (ID).

Specify the ID by [parameter](transportRequestUploadSOLMAN#By-Step-Parameter) or [common pipeline environment](transportRequestUploadSOLMAN#Common-Pipeline-Environment).

## Return Value

The step `isChangeInDevelopment` returns a boolean value by setting the custom key
`custom.isChangeInDevelopment` of the common pipeline environment.

If the change document is in status `in development` the key's value is set to `true`.

If the change document is _**not**_ in status `in development` the key's value is set to `false`. Furthermore, `AbortException` is thrown and the pipeline fails.

```groovy
// pipeline script
  isChangeInDevelopment( script: this )
  ...
```

You can omit the exception by setting the configuration parameter `failIfStatusIsNotInDevelopment` to `false`.

```groovy
// pipeline script
  isChangeInDevelopment( script: this, failIfStatusIsNotInDevelopment: false )

  if(commonPipelineEnvironment.getValue( 'isChangeInDevelopment' ) {
    ...
  }
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

```yaml
# config.yaml
general:
  changeManagement:
    credentialsId: 'SOLMAN_CRED_ID'
    endpoint: 'https://example.org/cm/solman/endpoint'

steps:
  isChangeInDevelopment:
    dockerImage: 'ppiper/cm-client'

  transportRequestUploadSOLMAN:
    dockerImage: 'ppiper/cm-client'
    applicationId: 'APPID',
    filePath: '/path/file.ext',
```

```groovy
// pipeline script
   ...
   stage('Upload') {
      transportRequestDocIDFromGit( script: this )

      isChangeInDevelopment( script: this )

      transportRequestReqIDFromGit( script: this )
      transportRequestUploadSOLMAN( script: this )
   }
```
