# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have installed the SAP component SAP_UI 7.53 or higher on your ABAP system.
* You have enabled the OData Service to Load Data to the [SAPUI5 ABAP Repository](https://sapui5.hana.ondemand.com/#/topic/a883327a82ef4cc792f3c1e7b7a48de8.html)
* You have the S_DEVELOP authorization for operations on a SAPUI5 ABAP repository.
* You have created a transport request on the ABAP system, which is the target container of the SAPUI5 application to upload.

[SAP Fiori Tools User Guide](https://help.sap.com/viewer/17d50220bcd848aa854c9c182d65b699/Latest/en-US)
[SAP Fiori Tools Deployment](https://help.sap.com/viewer/17d50220bcd848aa854c9c182d65b699/Latest/en-US/1b7a3be8d99c45aead90528ef472af37.html)
[Generate Deployment Configuration ABAP](https://help.sap.com/viewer/17d50220bcd848aa854c9c182d65b699/Latest/en-US/c06b9cbb3f3641aabfe3a5d199e855a0.html)
[Fiori Deploy](https://www.npmjs.com/package/@sap/ux-ui5-tooling#fiori-deploy---performs-the-deployment-of-the-application-into-an-abap-system)

## Upload Client Setup

The step `transportRequestUploadCTS` uses the [Node.js](https://nodejs.org)-based [SAP Fiori tools](https://help.sap.com/viewer/product/SAP_FIORI_tools/Latest/en-US) to upload your SAPUI5 application to the UI5 repository service of your ABAP back-end infrastructure. It performs a deployment running the [Fiori deploy command](https://www.npmjs.com/package/@sap/ux-ui5-tooling#fiori-deploy---performs-the-deployment-of-the-application-into-an-abap-system) on a Docker image.
By default, a plain [Node.js Docker image](https://hub.docker.com/_/node) is pulled and equipped with the SAPUI5 toolset during runtime of the pipeline.
Alternatively, you can provide your own, fully equipped Docker image. This speeds-up the upload process, but obligates to maintain and provision the image on a Docker registry.

### Fully Equipped Docker image

To create an own Docker image with the [SAP Fiori tools](https://help.sap.com/viewer/product/SAP_FIORI_tools/Latest/en-US) installed, proceed as follows:

1. Create a Node-based Docker image with the SAP Fiori tools installed:

   ```Dockerfile
   FROM node
   USER root
   RUN npm install -global @ui5/cli @sap/ux-ui5-tooling @ui5/logger @ui5/fs
   USER node
   ```

   ```/bin/bash
   docker build -t my/fiori-node .
   ```

1. Push your image to your private [Docker Hub registry](https://hub.docker.com/).

   ```/bin/bash
   docker push my/fiori-node  
   ```

1. Add the following to your config.yml file:

```yaml
steps:
  transportRequestUploadCTS:
    dockerImage: 'my/fiori-node'
    deployToolDependencies: []
```

## Building the SAPUI5 Application

Build your SAPUI5 application with the build command of the SAPUI5 toolset and use the step [npmExecuteScripts](npmExecuteScripts.md) to run the build command, proceeding as follows:

1. Configure the step in the `package.json` file of your project

   ```json
   {
      ...
      "scripts": {
         "start": "ui5 serve",
         "test": "npm run lint",
         "build": "ui5 build -a --clean-dest",
         ...
      },
      "dependencies": {},
      "devDependencies": {
         "@ui5/cli": "^2.11.2",
         ...
      }
   }
   ```

1. Setup the execution step in the pipeline

   ```groovy
   stage('Build') {
      npmExecuteScripts(script: this, runScripts: ['build'])
   }
   ```

**Note:** Do not use the `mtaBuild` step. The `mta` is dedicated to the SAP Cloud Platform. It does neither create the expected `dist` folder nor the compliant content.

## Uploading the SAPUI5 Application

The Fiori toolset uses the [ODATA service](https://ui5.sap.com/#/topic/a883327a82ef4cc792f3c1e7b7a48de8) to load your UI5 application to the SAPUI5 ABAP Repository.
The access is controlled by Basic Authentication (user/password based authentication).

**Note:** Use an HTTPS endpoint to ensure the encryption of your credentials.

## Specifying the Transport Request

The target of the upload is a transport request, identified by an identifier (ID).

The step `transportRequestUploadCTS` allows you to set the ID by parameter.

Alternatively, you can pass the ID through the `commonPipelineEnvironment`.
For example, by performing a step that generates the ID or obtains it differently.
See [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md).

### Adding a Parameter

A parameterized pipeline allows you to specify the ID with the launch of the build instead of entering it statically into the pipeline.

```groovy
transportRequestUploadCTS(
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

transportRequestUploadCTS(
    script:this,
    transportRequestId: ids['TRANSPORT_REQUEST_ID'],
    ...
)
```

## Common Pipeline Environment

Use the step [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md) to obtain the  `transportRequestId` value from your Git commit messages.

This step extracts the ID from the commit messages of your project repository and enters it into the `commonPipelineEnvironment`, in turn, the upload step `transportRequestUploadCTS` picks it up from there.

```groovy
transportRequestReqIDFromGit( script: this )
transportRequestUploadCTS( script: this, ... )
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```yaml
# config.yaml
steps:
  transportRequestUploadCTS:
    changeManagement:
      credentialsId: 'CTS_CREDENTIALS_ID'
      endpoint: 'https://example.org/sap/opu/odata/SAP/SCTS_CLOUD_API_ODATA_SRV'
      client: '001'
    abapPackage: 'PACK'
    applicationName: 'APP'
    applicationDescription: 'Lorem ipsum'
```

```groovy
// pipeline script
   stage('Init') {
      transportRequestReqIDFromGit( script: this )
   }
   stage('Build') {
      npmExecuteScripts( script: this, runScripts: ['build'])
   }
   stage('Upload') {
      transportRequestUploadCTS( script: this)
   }
```
