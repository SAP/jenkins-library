# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

With this step, you can retrieve all the API providers from the API portal. An API provider is a concept in API Management, capability of SAP Integration Suite, which defines the connection details for services running on specific hosts whose details you want to access.

You use an API provider to define not only the details of the host you want an application to reach, but also to define any further details that are necessary to establish the connection, for example, proxy settings. For more details, see the blog [API Providers](https://blogs.sap.com/2017/07/27/blog-series-api-providers/)

To consume the ApiProviderList step, proceed as follows:

* Copy the SAP API management service key from the SAP BTP cockpit. Go to Instance and Subscriptions &rarr; service API Management, API portal, which was created under apiportal-apiaccess plan.
* Store your service key created for API Management in the Jenkins server as a secret text.
* Create a new Jenkins pipeline designated for the ApiProviderList step.
* Execute the pipeline and validate the step exection results as explained in the blog [Integration Suite Piper commands](https://blogs.sap.com/2022/01/05/orking-with-integration-suite-piper-commands/)
* Use the ApiProviderList step to get the api providers list from the API Portal.
* ApiProviderList only supports GET operation.

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Configuration example for a `Jenkinsfile`:

```groovy
apiProviderList script: this
```

Configuration example for a YAML file(for example `.pipeline/config.yaml`):

```yaml
steps:
  <...>
  apiProviderList:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    Top: MY_API_PROVIDER_GET_N_ENTITIES
    Skip: MY_API_PROVIDER_SKIP_N_ENTITIES
    Filter: MY_API_PROVIDER_FILTER_BY_ENTITY_FIELD
    Orderby: MY_API_PROVIDER_ORDER_BY_ENTITY_FIELD
    Count: MY_API_PROVIDER_ORDER_ENTITY_COUNT
    Search: MY_API_PROVIDER_SEARCH_BY_ENTITY_FIELD
    Select: MY_API_PROVIDER_SELECT_BY_ENTITY_FIELD
    Expand: MY_API_PROVIDER_EXPAND_BY_ENTITY_FIELD
```
