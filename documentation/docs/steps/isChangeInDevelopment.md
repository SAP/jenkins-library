# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* You have an SAP Solution Manager user to which you have assigned the roles required for uploading. See [SAP Solution Manager Administration](https://help.sap.com/viewer/c413647f87a54db59d18cb074ce3dafd/7.2.12/en-US/11505ddff03c4d74976dae648743e10e.html).
* You have created a change document.
* You have installed the Change Management Client with the needed certificates. See [Change Management Client](transportRequestUploadSOLMAN.md#change-management-client).

## Specifying the Change Document

The target of the status check is a change document identified by an identifier (ID).

Specify the ID by [step parameter](transportRequestUploadSOLMAN.md#by-step-parameter) or [common pipeline environment](transportRequestUploadSOLMAN.md#common-pipeline-environment).

## Return Value

The step `isChangeInDevelopment` returns a boolean value by setting the custom key
`custom.isChangeInDevelopment` of the common pipeline environment:

- `true` if the change document is in status `in development`.

- `false` if the change document is _**not**_ in status `in development`. In this case, `AbortException` terminates the execution of the pipeline job.

```groovy
// pipeline script
  isChangeInDevelopment( script: this )
  ...
```

You can omit this exception by setting the configuration parameter `failIfStatusIsNotInDevelopment` to `false`:

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
    dockerImage: 'ppiper/cm-client:3.0.0.0'

  transportRequestUploadSOLMAN:
    dockerImage: 'ppiper/cm-client:3.0.0.0'
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
