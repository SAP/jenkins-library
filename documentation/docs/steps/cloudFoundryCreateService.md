# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

The following Example will create the services specified in a file `manifest-create-service.yml` in cloud foundry org `cfOrg` of Cloud Foundry installation accessed via `https://test.server.com` in space `cfSpace` by using the username & password stored in `cfCredentialsId`.

```groovy
cloudFoundryCreateService(
            script: this,
            cloudFoundry: [apiEndpoint: 'https://test.server.com',
                credentialsId: 'cfCredentialsId',
                serviceManifest: 'manifest-create-service.yml',
                org: 'cfOrg',
                space: 'cfSpace'])
```

The following example additionally to above also makes use of a variable substitution file `mainfest-variable-substitution.yml`.

```groovy
cloudFoundryCreateService(
            script: this,
            cloudFoundry: [apiEndpoint: 'https://test.server.com',
                credentialsId: 'cfCredentialsId',
                serviceManifest: 'manifest-create-service.yml',
                manifestVariablesFiles: ['mainfest-variable-substitution.yml'],
                org: 'cfOrg',
                space: 'cfSpace'])

```
