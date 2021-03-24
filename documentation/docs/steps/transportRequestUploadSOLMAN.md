# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* **Solution Manager Account** - user account with according roles for upload
* **Change Document** - the change document the tranport request is coupled to
* **Transport Request** - the target container of the upload
* **Change Management Client** - installation of the CM Client with needed certificates  

## Change Management Client

Access to the Solution Manager is handled via the so-called CM Client.
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
${docGenStepName} allows to set these IDs by parameter or to use Git commit messages.
As an additional option, IDs can be passed in via the Common Pipeline Environment. For example through a
step that generates the IDs or obtains them differently.

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
        string(name: 'CHANGE_DOCUMENT_ID', description: 'Change Document ID'), 
        string(name: 'TRANSPORT_REQUEST_ID', description: 'Transport Request ID')
    ]
)

transportRequestUploadSOLMAN(
    script:this, 
    changeDocumentId: ids['CHANGE_DOCUMENT_ID'], 
    transportRequestId: ids['TRANSPORT_REQUEST_ID'],
    ...
)
```

### By Git Commit Message

If the identifiers are neither defined as step parameters nor by the Common Pipeline Environment,
the Git commit messages (`git log`) of the project are searched for lines that follow a defined pattern.
The pattern is specified by the label _changeDocumentLabel_ (default=`ChangeDocument`) resp.
_transportRequestLabel_ (default=`TransportRequest`). Behind the label a colon
any blanks and the identifier are expected.

```
Release - define IDs for upload to Solution Manager

   ChangeDocument: 1000001234
   TransportRequest: ABCD10005E
```

The IDs dont need to be defined in the same message.

The Git commit messages to be considered are determined by the parameters _gitFrom_ (default=`origin/master`)
and _gitTo_ (default=`HEAD`). The naming follows the Git revision range representation `git log <gitFrom>..<gitTo>`.
All commit messages accessible from _gitTo_ but not from _gitFrom_ are taken into account. If the scanner
detects multiple IDs, it fails. So the commit range has to be chosen accordingly.

In case of a pull request of a feature branch, the default should be sufficient as long as the transport request
isn't changed. Only the commits (`HEAD`) that have not yet entered the main branch `origin/master` would be scanned.

If uploading from the main branch, it must be assumed that former change document and transport request IDs
are already contained in the history. In this case the new IDs should be maintained in the `HEAD` and
_gitFrom_ be set to `HEAD~1`.

```yaml
general:
  changeManagement:
    git:
      from: 'HEAD~1'
```

## Common Pipeline Environment

Werden `changeDocumentId` und `transportRequestId` erst zur Laufzeit über _Git commit messages_ ermittelt, so werden diese in das `commonPipelineEnvironment` eingetragen und sind entsprechend arufbar.

```
  this.commonPipelineEnvironment.getValue('changeDocumentID')
  this.commonPipelineEnvironment.getValue('transportRequestID')
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
  changeDocumentId: '12345678',
  transportRequestId: '87654321',
  filePath: '/path/file.ext',
  cmClientOpts: '-Dkey=value'
)
```
