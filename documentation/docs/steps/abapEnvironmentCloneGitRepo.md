# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

A SAP BTP, ABAP environment system is available.
On this system, a [Communication User](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/0377adea0401467f939827242c1f4014.html), a [Communication System](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/1bfe32ae08074b7186e375ab425fb114.html) and a [Communication Arrangement](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/a0771f6765f54e1c8193ad8582a32edb.html) is setup for the Communication Scenario "Software Component Management Integration (SAP_COM_0948)". This can be done manually through the respective applications on the SAP BTP, ABAP environment system or through creating a service key for the system on Cloud Foundry with the parameters {"scenario_id": "SAP_COM_0948", "type": "basic"}. In a pipeline, you can do this with the step [cloudFoundryCreateServiceKey](https://sap.github.io/jenkins-library/steps/cloudFoundryCreateServiceKey/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example: Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapEnvironmentCheckoutBranch script: this
```

If you want to provide the host and credentials of the Communication Arrangement directly, the configuration could look as follows:

```yaml
steps:
  abapEnvironmentCloneGitRepo:
    repositoryName: '/DMO/GIT_REPOSITORY'
    branchName: 'my-demo-branch'
    abapCredentialsId: 'abapCredentialsId'
    host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
```

Please note that the branchName parameter specifies the target branch you want to clone. Also keep in mind that the repositoryName parameter must define a single repository.

Another option is to read the host and credentials from the cloud foundry service key of the respective instance. Furthermore, if you want to clone multiple repositories, they can be specified in a configuration file.

With this approach the `config.yml` would look like this:

```yaml
steps:
  abapEnvironmentCloneGitRepo:
    repositories: 'repositories.yml'
    cfCredentialsId: 'cfCredentialsId'
    cfApiEndpoint: 'https://test.server.com'
    cfOrg: 'cfOrg'
    cfSpace: 'cfSpace'
    cfServiceInstance: 'cfServiceInstance'
    cfServiceKeyName: 'cfServiceKeyName'
```

and the configuration file `repositories.yml` would look like this:

```yaml
repositories:
  - name: '/DMO/REPO'
    branch: 'main'
  - name: '/DMO/REPO_COMMIT'
    branch: 'feature'
    commitID: 'cd87a3cac2bc946b7629580e58598c3db56a26f8'
  - name: '/DMO/REPO_TAG'
    branch: 'release'
    tag: 'myTag'
```

Using such a configuration file is the recommended approach. Please note that you need to use the YAML data structure as in the example above when using the `repositories.yml` config file.
If you want to clone a specific commit, either a `commitID` or a `tag` can be specified. If both are specified, the `tag` will be ignored.

## Example: Configuration in the Jenkinsfile

It is also possible to call the steps - including all parameters - directly in the Jenkinsfile.
In the first example, the host and the credentialsId of the Communication Arrangement are directly provided.

```groovy
abapEnvironmentCloneGitRepo (
  script: this,
  repositoryName: '/DMO/GIT_REPOSITORY',
  branchName: 'my-demo-branch',
  abapCredentialsId: 'abapCredentialsId',
  host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
)
```

In the second example, the host and credentialsId will be read from the provided cloud foundry service key of the specified service instance.

```groovy
abapEnvironmentCloneGitRepo (
  script: this,
  repositoryName: '/DMO/GIT_REPOSITORY',
  branchName: 'my-demo-branch'
  abapCredentialsId: 'cfCredentialsId',
  cfApiEndpoint: 'https://test.server.com',
  cfOrg: 'cfOrg',
  cfSpace: 'cfSpace',
  cfServiceInstance: 'cfServiceInstance',
  cfServiceKeyName: 'cfServiceKeyName'
)
```

## Example: Cloning a Bring Your Own Git (BYOG) repository

> Feature will be available in November 2024.

Since a ByoG repository is an external repository, you must be authenticated to clone it.
For this, the corresponding credentials must be stored in Jenkins as a username and password/token.

<strong> Store the credentials: </strong> <br>
A new credential with the type username and password must be stored.<br>
`Jenkins Dashboard > Manage Jenkins > Credentials` <br>
These credentials are used to clone the ByoG repository.
More information on configuring the credentials can be found [here](https://www.jenkins.io/doc/book/using/using-credentials/).

The config.yaml should look like this:

```yaml
steps:
  abapEnvironmentCloneGitRepo:
    repositories: 'repos.yaml'
    byogCredentialsId: 'byogCredentialsId'
    abapCredentialsId: 'abapCredentialsId'
    host: '1234-abcd-5678-efgh-ijk.abap.eu10.hana.ondemand.com'
```

`byogCredentialsId: 'byogCredentialsId'` is the reference to the defined credential in Jenkins. So take care that this matches with your setup.

After that, the ByoG repository that is to be cloned must be specified in the repos.yaml:

```yaml
repositories:
  - name: '/DMO/REPO_BYOG'
    branch: 'main'
```

After the pipeline has run through, the repository has been cloned.
