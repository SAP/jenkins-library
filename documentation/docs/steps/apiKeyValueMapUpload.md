# ${docGenStepName}

## ${docGenDescription}

With this step you can store one or more key value pairs of data stored in a group called a map or key value map.

To consume the ApiKeyValueMapUpload step, proceed as follows:

* Copy the SAP API management service key from the SAP BTP sub account cockpit, under instance and subscriptions &rarr; service API Management, API portal, which was created under apiportal-apiaccess plan.
* Store your service key created for SAP API Management in the Jenkins server as a secret text.
* Create a new Jenkins pipeline designated for the ApiKeyValueMapUpload step.
* Execute the pipeline and validate the step exection results as explained in the blog [Integration Suite Piper commands](https://blogs.sap.com/2022/01/05/orking-with-integration-suite-piper-commands/)
* Using the ApiKeyValueMapUpload step, you can create a new API key value map in the API portal.
* The ApiKeyValueMapUpload step allows you to prevent command execution in case an API key value map already exists.
* If API key value map already exists, then delete it and execute the piper step again, which will create a new API Key value Map.
* ApiKeyValueMapUpload only supports create operation.

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Configuration example for a `Jenkinsfile`:

```groovy
apiKeyValueMapUpload script: this
```

Configuration example for a YAML file(for example `.pipeline/config.yaml`):

```yaml
steps:
  <...>
  apiKeyValueMapUpload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    key: API_KEY_NAME
    value: API_KEY_VALUE
    keyValueMapName: API_KEY_VALUE_MAP_NAME
```
