# ${docGenStepName}

## ${docGenDescription}

With this step, you can add a new entity to the API Providers. An API Provider is a concept in API Management, capability of SAP Integration Suite, which defines the connection details for services running on specific hosts whose details you want to access.

You use an API provider to define not only the details of the host you want an application to reach, but also to define any further details that are necessary to establish the connection, for example, proxy settings.For more details, see the blog [API Providers](https://blogs.sap.com/2017/07/27/blog-series-api-providers/)

To consume the ApiProviderUpload step, proceed as follows:

* Copy the SAP API management service key from the SAP BTP cockpit. Go to instance and subscriptions &rarr; service API Management, API portal, which was created under apiportal-apiaccess plan.
* Store your service key created for API Management in the Jenkins server as a secret text.
* Create a new Jenkins pipeline designated for the ApiKeyValueMapUpload step.
* Create a api provider json file in the jenkins worksapce relative file path, as an input for ApiKeyValueMapUpload step.
* Execute the pipeline and validate the step exection results as explained in the blog [Integration Suite Piper commands](https://blogs.sap.com/2022/01/05/orking-with-integration-suite-piper-commands/)
* Use the ApiProviderUpload step to create a new API provider in the API portal.
* If API provider already exists, then delete it and execute the piper step again, which will create a new API provider.
* ApiProviderUpload only supports create operation.

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Configuration example for a `Jenkinsfile`:

```groovy
apiProviderUpload script: this
```

Configuration example for a YAML file(for example `.pipeline/config.yaml`):

```yaml
steps:
  <...>
  apiKeyValueMapUpload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    filePath: MY_API_PROVIDER_JSON_FILE_PATH
```
