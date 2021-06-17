# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have created an RFC destination on ABAP system side and registered an RFC server.
* You have an user account on the ABAB system and the roles required for uploading via RFC.
* You have created a transport request on the ABAB backend, which is the target container of the upload.
* You have installed an RFC library based Connector Client. See [RFC Client](#[RFC Client).

## RFC Client

Create a Docker image as described in the Git repository [devops-docker-images/node-rfc](https://github.com/rodibrin/devops-docker-images/tree/master/node-rfc).
Put the image in its own Docker register in [Docker Hub](https://hub.docker.com/).
Make the image known to the pipeline in the configuration.

```yaml
steps:
  transportRequestUploadRFC:
    dockerImage: '<my>/rfc-client'
```

The RFC Client connects to the ABAP system using the [SAP NetWeaver RFC SDK](https://support.sap.com/en/product/connectors/nwrfcsdk.html).
See the documentation of the [classical SAP connectivity technology RFC](https://help.sap.com/viewer/753088fc00704d0a80e7fbd6803c8adb/1709%20000/en-US/4888068ad9134076e10000000a42189d.html) for detailed information.

## Specifying the Transport Request

The target of the upload is a transport request, identified by an identifier (ID).
The step `transportRequestUploadRFC` allows you to set the ID by parameter.
Alternatively, you can pass the ID through the Common Pipeline Environment.
For example, by performing a step that generates the ID or obtains it differently.
See [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md)

### By Step Parameter

A parameterized pipeline allows to specify the ID with the launch of the build
instead of entering it statically into the pipeline.

```groovy
transportRequestUploadRFC(
    script: this,
    transportRequestId: ${TRANSPORT_REQUEST_ID},
    ...
)
```

The Jenkins pipeline `input` step allows to specify the ID at runtime of the pipeline.

```groovy
def ids = input( message: "Upload?",
    parameters: [
        string(name: 'TRANSPORT_REQUEST_ID',description: 'Transport Request ID')
    ]
)

transportRequestUploadRFC(
    script:this,
    transportRequestId: ids['TRANSPORT_REQUEST_ID'],
    ...
)
```

## Common Pipeline Environment

With OS Piper you can use the step [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md) to obtain the  `transportRequestId` value from your Git commit messages.
The step enter the ID into the `commonPipelineEnvironment`, in turn, the upload step `transportRequestUploadRFC` picks it up from there.

```groovy
transportRequestReqIDFromGit( script: this )
transportRequestUploadRFC( script: this, ... )
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```yaml
steps:
  transportRequestUploadRFC:
    changeManagement:
      credentialsId: 'RFC_CREDENTIALS_ID'
      endpoint: '<Application server URL>'
      instance: '00'
      client: '001'
    abapPackage: ''
    applicationDescription: ''
    applicationName: ''
    dockerImage: '<my>/rfc-client'
```

```groovy
transportRequestUploadRFC(
  script: this,
                transportRequestId: "A5DK000085",
                applicationUrl: 'https://nexussnap.wdf.sap.corp:8443/nexus/content/repositories/deploy.snapshots/com/sap/marcusholl/1.0-SNAPSHOT/archive.zip'
  endpoint: 'https://example.org/cm/rfc/endpoint'
  applicationId: 'ABC',
  uploadCredentialsId: "RFC_CREDENTIALS_ID"
  changeDocumentId: '1000001234',
  transportRequestId: 'ABCD10005E',
  filePath: '/path/file.ext',
  cmClientOpts: '-Dkey=value'
)
```
