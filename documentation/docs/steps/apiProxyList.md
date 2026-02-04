# ${docGenStepName}

## ${docGenDescription}

With this step, you can retrieve all the API proxies from the API portal. An API Proxy is a concept in API Management, capability of SAP Integration Suite, which anonymizes any HTTP endpoints like REST, OData, or SOAP and enhance it with policies and routes.

An API proxy is a discrete representation of an API. It is implemented as a set of configuration files, policies, and code snippets that rely on the resource information provided by API Management. For more information, see the document [API Proxy](https://help.sap.com/doc/654e5912ee554d46bcc6347599fb2096/CLOUD/en-US/Unit%2004.3%20-%20API%20Proxy%20-%20API%20Resource.pdf/)

To consume the ApiProxyList step, proceed as follows:

* Copy the SAP API management service key from the SAP BTP cockpit. Go to instance and subscriptions &rarr; service API Management, API portal, which was created under apiportal-apiaccess plan.
* Store your service key created for API Management in the Jenkins server as a secret text.
* Create a new Jenkins pipeline designated for the ApiProxyList step.
* Execute the pipeline and validate the step exection results as explained in the blog [Integration Suite Piper commands](https://blogs.sap.com/2022/01/05/orking-with-integration-suite-piper-commands/)
* Use the ApiProxyList step to get the api proxy list from the API portal.
* ApiProxyList only supports GET operation.

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Configuration example for a `Jenkinsfile`:

```groovy
apiProxyList script: this
```

Configuration example for a YAML file(for example `.pipeline/config.yaml`):

```yaml
steps:
  <...>
  apiProxyList:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    Top: MY_API_PROXY_GET_N_ENTITIES
    Skip: MY_API_PROXY_SKIP_N_ENTITIES
    Filter: MY_API_PROXY_FILTER_BY_ENTITY_FIELD
    Orderby: MY_API_PROXY_ORDER_BY_ENTITY_FIELD
    Count: MY_API_PROXY_ORDER_ENTITY_COUNT
    Search: MY_API_PROXY_SEARCH_BY_ENTITY_FIELD
    Select: MY_API_PROXY_SELECT_BY_ENTITY_FIELD
    Expand: MY_API_PROXY_EXPAND_BY_ENTITY_FIELD
```
