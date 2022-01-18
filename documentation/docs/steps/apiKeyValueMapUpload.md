# ${docGenStepName}

## ${docGenDescription}

* ApiKeyValueMapUpload stores one or more key value pairs of data in a grouping called a map, or key value map.

* A typical flow would look something like

You have to store API management service key in jenkins server as secret text. this is obtained from the SAP BTP sub account cockpit, under instance and subscriptions for API management under API plan. then you create new jenkins pipeline, which consumes the ApiKeyValueMapUpload piper step. you can execute the pipeline and validate step exection results as explained in the blog [Integration Suite Piper commands](https://blogs.sap.com/2022/01/05/working-with-integration-suite-piper-commands/)

* With the ApiKeyValueMapUpload step

You can create API key value map in the API portal.

If API key value map already exist, then command execution would fail.

If API key value map already exist, then delete it and execute the piper step again,
which create new API Key value Map.

ApiKeyValueMapUpload only supports create operation, but not either delete, get, update, which is supported in different piper steps.

## Prerequisites

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
apiKeyValueMapUpload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  apiKeyValueMapUpload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    key: API_KEY_NAME
    value: API_KEY_VALUE
    keyValueMapName: API_KEY_VALUE_MAP_NAME
```
