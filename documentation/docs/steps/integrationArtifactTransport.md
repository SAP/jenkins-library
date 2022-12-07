# ${docGenStepName}

## ${docGenDescription}

With this step, you can transport Cloud Integration, capability of SAP Integration Suite content across various landscapes using SAP Content Agent Service.

Cloud Integration provides content transport mechanism. SAP Content Agent service enables you to assemble the content from these content providers in MTAR format. Later, this content is either available for download or can be exported to the configured transport queue, such as SAP Cloud Transport Management. For more information on
configurations required for Cloud Integration, see [Content Assembly for SAP Integration Suite](https://help.sap.com/docs/CONTENT_AGENT_SERVICE/ae1a4f2d150d468d9ff56e13f9898e07/8e274fdd41da45a69ff919c0af8c6127.html)

To use the integrationArtifactTransport step, proceed as follows:

* [Create SAP Content Agent Service Destination](https://help.sap.com/docs/CONTENT_AGENT_SERVICE/ae1a4f2d150d468d9ff56e13f9898e07/a4da0c26ced74bbfbc60e7f607dc05ab.html).
* [Create SAP Cloud Integration Destination](https://help.sap.com/docs/CONTENT_AGENT_SERVICE/ae1a4f2d150d468d9ff56e13f9898e07/c17c4004049d4d9dba373d72ce5610cd.html).
* [Create SAP Cloud Transport Management Destination](https://help.sap.com/docs/CONTENT_AGENT_SERVICE/ae1a4f2d150d468d9ff56e13f9898e07/b44463a657fa4be48ea2525b7eb6e7de.html).
* Transport SAP Cloud Integration Content with CAS as explained in the blog [Transport SAP Cloud Integration (CI/CPI) Content with Transport Management Service (TMS) and Content Agent Service (CAS)](https://blogs.sap.com/2022/03/25/transport-sap-cloud-integration-ci-cpi-content-with-transport-management-service-tms-and-content-agent-service-cas/)
* integrationArtifactTransport step only supports Integration Package transport.

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Configuration example for a `Jenkinsfile`:

```groovy
integrationArtifactTransport script: this
```

Configuration example for a YAML file(for example `.pipeline/config.yaml`):

```yaml
steps:
  <...>
  integrationArtifactTransport:
    casApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationPackageId: MY_INTEGRATION_PACKAGE_ID
    resourceID: MY_INTEGRATION_RESOURCE_ID
    name: MY_INTEGRATION_PACKAGE_NAME
    version: MY_INTEGRATION_PACKAGE_VERSION
```
