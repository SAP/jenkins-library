# ${docGenStepName}

## ${docGenDescription}

## Prerequisites
* ???

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentAssembly script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentAssembly:
    abapCredentialsId: 'abapCredentialsId',
    host: 'https://myABAPendpoint.com',
```

Or by authenticating against Cloud Foundry and reading the Service Key details from there:

```yaml
steps:
  abapEnvironmentAssembly:
    abapCredentialsId: 'cfCredentialsId',
    cfApiEndpoint : 'https://test.server.com',
    cfOrg : 'cfOrg',
    cfSpace: 'cfSpace',
    cfServiceInstance: 'myServiceInstance',
```
