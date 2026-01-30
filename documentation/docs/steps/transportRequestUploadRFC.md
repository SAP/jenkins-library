# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* You have enabled RFC on the ABAP system.
* You have a user account on the ABAP system where you have assigned the required roles for uploading via RFC.
* You have created a transport request on the ABAP system, which is the target container of the upload.

## RFC Client

The RFC Client connects to your ABAP system using the [SAP NetWeaver RFC SDK](https://support.sap.com/en/product/connectors/nwrfcsdk.html).

For more information, see [classical SAP connectivity technology RFC](https://help.sap.com/viewer/753088fc00704d0a80e7fbd6803c8adb/1709%20000/en-US/4888068ad9134076e10000000a42189d.html).

To install an RFC library based Connector Client, proceed as follows:

1. Create a Docker image as described in the Git repository [devops-docker-images/node-rfc](https://github.com/rodibrin/devops-docker-images/tree/master/node-rfc).
1. Push your image to your private [Docker Hub registry](https://hub.docker.com/).
1. Add the following to your config.yml file:

```yaml
steps:
  transportRequestUploadRFC:
    dockerImage: 'my/rfc-client'
```

## Specifying the Transport Request

The target of the upload is a transport request, identified by an identifier (ID).

The step `transportRequestUploadRFC` allows you to set the ID by parameter.

Alternatively, you can pass the ID through the `commonPipelineEnvironment`.
For example, by performing a step that generates the ID or obtains it differently.
See [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md).

### Adding a Parameter

A parameterized pipeline allows you to specify the ID with the launch of the build instead of entering it statically into the pipeline.

```groovy
transportRequestUploadRFC(
    script: this,
    transportRequestId: ${TRANSPORT_REQUEST_ID},
    ...
)
```

The Jenkins pipeline `input` step allows you to specify the ID at runtime of the pipeline.

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

Use the step [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md) to obtain the  `transportRequestId` value from your Git commit messages.

This step extracts the ID from the commit messages of your project repository and enters it into the `commonPipelineEnvironment`, in turn, the upload step `transportRequestUploadRFC` picks it up from there.

```groovy
transportRequestReqIDFromGit( script: this )
transportRequestUploadRFC( script: this, ... )
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```yaml
# config.yaml
steps:
  transportRequestUploadRFC:
    changeManagement:
      credentialsId: 'RFC_CREDENTIALS_ID'
      endpoint: 'https://example.org/cm/rfc/endpoint'
      instance: '00'
      client: '001'
    abapPackage: 'PACK'
    applicationDescription: 'Lorem ipsum'
    applicationName: 'APP'
    dockerImage: 'my/rfc-client'
```

```groovy
// pipeline script
transportRequestReqIDFromGit( script: this )
transportRequestUploadRFC( script: this, applicationUrl: 'https://example.org/appl/url/archive.zip')
```
