# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

Example configuration for the use in a `Jenkinsfile`.

```groovy
integrationPackageDownload script: this
```

Example for the use in a YAML configuration file (such as `.pipeline/config.yaml`).

```yaml
steps:
  <...>
  integrationPackageDownload:
    cpiApiServiceKeyCredentialsId: 'MY_API_SERVICE_KEY'
    integrationPackageId: 'MY_INTEGRATION_PACKAGE_NAME'
    downloadPath: MY_INTEGRATION_PACKAGE_DOWNLOAD_PATH
```
