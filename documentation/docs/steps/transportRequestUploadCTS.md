# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* You have installed the [SAP component SAP_UI 7.53](https://help.sap.com/viewer/6f3c61a7a5b94447b80e72f722b0aad7/202009.002/en-US/35828457ed26452db8d51c840813f1bb.html) or higher on your ABAP system.
* You have enabled the OData Service to load data to the [SAPUI5 ABAP repository](https://sapui5.hana.ondemand.com/#/topic/a883327a82ef4cc792f3c1e7b7a48de8.html).
* You have the [S_DEVELOP authorization](https://sapui5.hana.ondemand.com/#/topic/a883327a82ef4cc792f3c1e7b7a48de8.html) to perform operations in your SAPUI5 ABAP repository.
* You have created a transport request on the ABAP system, which is the target container of the SAPUI5 application for your upload.

## Setting Up an Upload Client

The step `transportRequestUploadCTS` uses the [Node.js](https://nodejs.org)-based [SAP Fiori tools](https://help.sap.com/viewer/product/SAP_FIORI_tools/Latest/en-US) to upload your SAPUI5 application to the UI5 repository service of your ABAP back-end infrastructure. It performs the deployment command [Fiori deploy](https://www.npmjs.com/package/@sap/ux-ui5-tooling#fiori-deploy---performs-the-deployment-of-the-application-into-an-abap-system) on a Docker image.

By default, a plain [node.js Docker image](https://hub.docker.com/_/node) is pulled and equipped with the SAPUI5 toolset during runtime of the pipeline.
Alternatively, you can provide your own, fully equipped Docker image. This speeds up the upload process, but requires you to maintain and provision the image on a Docker registry.

### Creating a Fully Equipped Docker Image

To create an own Docker image with the [SAP Fiori tools](https://help.sap.com/viewer/product/SAP_FIORI_tools/Latest/en-US), proceed as follows:

1. Create a node.js based Docker image with the SAP Fiori tools installed:

    ```Dockerfile
    FROM node
    USER root
    RUN npm install -global @ui5/cli @sap/ux-ui5-tooling @ui5/logger @ui5/fs
    USER node
    ```

    ```/bin/bash
    docker build -t my/fiori-node .
    ```

1. Push your image to your private [Docker Hub registry](https://hub.docker.com/):

    ```/bin/bash
    docker push my/fiori-node
    ```

1. Add the following content to your `config.yml` file:

    ```yaml
    steps:
      transportRequestUploadCTS:
        dockerImage: 'my/fiori-node'
        deployToolDependencies: []
    ```

## Building an SAPUI5 Application

Build your SAPUI5 application with the build command of the SAPUI5 toolset and use the step [npmExecuteScripts](npmExecuteScripts.md) to run the build command. Proceed as follows to do so:

1. Configure the steps in the `package.json` file of your project as follows:

    ```json
    {
       ...
       "scripts": {
          "start": "ui5 serve",
          "test": "npm run lint",
          "build": "ui5 build --clean-dest",
          ...
       },
       "dependencies": {},
       "devDependencies": {
          "@ui5/cli": "^2.11.2",
          ...
       }
    }
    ```

1. Configure the execution step in the pipeline as follows:

    ```groovy
    stage('Build') {
       npmExecuteScripts(script: this, runScripts: ['build'])
    }
    ```

**Note:** Do not use the `mtaBuild` step. The MTA Build Tool `mta` is dedicated to the SAP Business Technology Platform. It does neither create the expected `dist` folder nor the compliant content.

## Uploading an SAPUI5 Application

The Fiori toolset uses the [ODATA service](https://ui5.sap.com/#/topic/a883327a82ef4cc792f3c1e7b7a48de8) to upload your UI5 application to the SAPUI5 ABAP repository. It controls access by [Basic Authentication](https://help.sap.com/viewer/e815bb97839a4d83be6c4fca48ee5777/202009.002/en-US/43960f4a527b58c1e10000000a422035.html?q=basic%20authentication) (user/password based authentication).

**Note:** Do not upload your application to SAP Business Technology Platform. The SAP BTP does not support `Basic Authentication`.

**Note:** Use an HTTPS endpoint to ensure the encryption of your credentials.

## Specifying the Transport Request

The target of the upload is a transport request, identified by an identifier (ID).

The step `transportRequestUploadCTS` allows you to set the ID by parameter.

Alternatively, you can pass the ID through the parameter `commonPipelineEnvironment`.
For example, by performing a step that generates the ID or obtains it differently.
For more information, see [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md).

### Adding a Parameter

A parameterized pipeline allows you to specify the ID with the launch of each build instead of entering it statically into the pipeline.

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

This step extracts the ID from the commit messages of your project repository and enters it into the `commonPipelineEnvironment`. In turn, the upload step `transportRequestUploadCTS` picks it up from there.

```groovy
transportRequestReqIDFromGit( script: this )
transportRequestUploadCTS( script: this, ... )
```

## General Purpose Pipeline Release Stage

The step can also be configured via General Purpose Pipeline in Release stage using the config.yml as follows:

```groovy
stages:
  Release:
    transportRequestUploadCTS: true
```

This will initialize the step within the Release stage of the pipeline and will upload the desired application (SAPUI5/OPENUI5) to the SAPUI5 ABAP repository.

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
      endpoint: 'https://example.org'
      client: '001'
    abapPackage: 'PACK'
    applicationName: 'APP'
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
