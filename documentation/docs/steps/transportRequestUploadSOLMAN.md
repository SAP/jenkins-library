# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* You have an SAP Solution Manager user account and the roles required for uploading. See [SAP Solution Manager Administration](https://help.sap.com/viewer/c413647f87a54db59d18cb074ce3dafd/7.2.12/en-US/11505ddff03c4d74976dae648743e10e.html).
* You have a change document to which your transport request is coupled.
* You have a transport request, which is the target container of the upload.
* You have installed the Change Management Client with the needed certificates. See [Change Management Client](#change-management-client).

## Change Management Client

The Change Management Client (CM Client) handles the access to SAP Solution Manager.
The CM Client is a software running under Linux, which can initiate basic change management tasks
in the Solution Manager. The client is used by default
as a [Docker image](https://hub.docker.com/r/ppiper/cm-client),
but can also be installed as a [command line tool](https://github.com/SAP/devops-cm-client).

!!! note "Certificates"
    It is expected that the Solution Manager endpoint is secured by SSL and sends a certificate accordingly.
    The CM Client verifies the certificate. If the publisher of this certificate is unknown, the connection will be rejected.
    The CM Client uses the underlying JDK procedures for the verification.
    Accordingly, the issuer must be specified in the truststore of the JDK.

Create a clone of the image and add the necessary certificate to its truststore in case you use the [Docker image](https://hub.docker.com/r/ppiper/cm-client).
Extend the truststore of the environment with the necessary certificate in the case you use the immediate [command line tool](https://github.com/SAP/devops-cm-client).

## Specifying the Change Document and Transport Request

The target of the upload is a transport request and the associated change document.
Both objects are identified by identifiers (ID).
`transportRequestUploadSOLMAN` allows you to set IDs by parameter.
Alternatively, you can pass the IDs through the Common Pipeline Environment.
For example, by performing a step that generates the IDs or obtains them differently.
See [transportRequestDocIDFromGit](transportRequestDocIDFromGit.md) and [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md)

### By Step Parameter

A parameterized pipeline allows to specify the IDs with the launch of the build
instead of entering them statically into the pipeline.

```groovy
transportRequestUploadSOLMAN(
    script: this,
    changeDocumentId: ${CHANGE_DOCUMENT_ID},
    transportRequestId: ${TRANSPORT_REQUEST_ID},
    ...
)
```

The Jenkins pipeline `input` step allows to specify the IDs at runtime of the pipeline.

```groovy
def ids = input( message: "Upload?",
    parameters: [
        string(name: 'CHANGE_DOCUMENT_ID',description: 'Change Document ID'),
        string(name: 'TRANSPORT_REQUEST_ID',description: 'Transport Request ID')
    ]
)

transportRequestUploadSOLMAN(
    script:this,
    changeDocumentId: ids['CHANGE_DOCUMENT_ID'],
    transportRequestId: ids['TRANSPORT_REQUEST_ID'],
    ...
)
```

## Common Pipeline Environment

With OS Piper you can use the steps [transportRequestDocIDFromGit](transportRequestDocIDFromGit.md) and [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md) to obtain the `changeDocumentId` and `transportRequestId` values from your Git commit messages.
The steps enter the IDs into the `commonPipelineEnvironment`, in turn, the upload step `transportRequestUploadSOLMAN` picks them up from there.

```groovy
transportRequestDocIDFromGit( script: this )
transportRequestReqIDFromGit( script: this )
transportRequestUploadSOLMAN( script: this, ... )
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
transportRequestUploadSOLMAN(
  script: this,
  endpoint: 'https://example.org/cm/solman/endpoint'
  applicationId: 'ABC',
  uploadCredentialsId: "SOLMAN_CRED_ID"
  changeDocumentId: '1000001234',
  transportRequestId: 'ABCD10005E',
  filePath: '/path/file.ext',
  cmClientOpts: '-Dkey=value'
)
```
