# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
apiProxyDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  apiProxyDownload:
    apimApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    apiProxyName: 'MY_API_PROXY_NAME'
    downloadPath: MY_API_PROXY_DOWNLOAD_PATH
```
