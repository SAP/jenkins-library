# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have an SAP Solution Manager user account and the roles required for uploading. See [SAP Solution Manager Administration](https://help.sap.com/viewer/c413647f87a54db59d18cb074ce3dafd/7.2.12/en-US/11505ddff03c4d74976dae648743e10e.html).
* You have a change document to which your transport request is coupled.
* You have a transport request, which is the target container of the upload.
* You have installed the Change Management Client with the needed certificates. See [Change Management Client](#Change Management Client).

## Change Management Client

The Change Management Client (CM Client) handles the access to SAP Solution Manager.
The CM Client is a software running under Linux, which can initiate basic change management tasks
in the Solution Manager as well as in the CTS. The client is used by default
as  a [Docker image](https://hub.docker.com/r/ppiper/cm-client),
but can also be installed as a [command line tool](https://github.com/SAP/devops-cm-client).

### Certificates

It is expected that the Solution Manager endpoint is secured by SSL and sends a certificate accordingly.
The certificate is verified by the CM Client. If the publisher of this certificate is unknown,
the connection will be rejected. The client implemented in Java uses
the underlying JDK procedures for the verification. Accordingly, the issuer must be specified in the
truststore of the JDK. In the case of the [Docker image](https://hub.docker.com/r/ppiper/cm-client)
a clone of the image must be created with the necessary certificate added to its truststore.
In the case of the immediate [command line tool](https://github.com/SAP/devops-cm-client),
only the truststore of the environment needs to be extended.

## Specifying the Change document and transport request

The target of the upload is a Transport Request, which is determined by the identifiers (ID)
of the Request and the associated Change Document.
`transportRequestUploadSOLMAN` allows to set these IDs by parameter.
As an additional option, IDs can be passed in via the Common Pipeline Environment.
For example through a step that generates the IDs or obtains them differently.

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

OS Piper provides the steps [transportRequestDocIDFromGit](transportRequestDocIDFromGit.md) and [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md) to get `changeDocumentId` and `transportRequestId` from the Git commit messages.
The IDs are entered into the `commonPipelineEnvironment` and picked up there by the upload step.

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
  changeDocumentId: '12345678',
  transportRequestId: '87654321',
  filePath: '/path/file.ext',
  cmClientOpts: '-Dkey=value'
)
```
