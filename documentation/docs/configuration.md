# Configuration

Configure your project through a yml-file, which is located at `.pipeline/config.yml` in the **master branch** of your source code repository.

Your configuration inherits from the default configuration located at [https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml](https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml).

!!! caution "Adding custom parameters"
    Please note that adding custom parameters to the configuration is at your own risk.
    We may introduce new parameters at any time which may clash with your custom parameters.

Configuration of the project "Piper" steps as well as project "Piper" templates can be done in a hierarchical manner.

1. Directly passed step parameters will always take precedence over other configuration values and defaults
1. Stage configuration parameters define a Jenkins pipeline stage dependent set of parameters (e.g. deployment options for the `Acceptance` stage)
1. Step configuration defines how steps behave in general (e.g. step `cloudFoundryDeploy`)
1. General configuration parameters define parameters which are available across step boundaries
1. Custom default configuration provided by the user through a reference in the `customDefaults` parameter of the project configuration
1. Default configuration comes with the project "Piper" library and is always available

![Piper Configuration](images/piper_config.png)

## Collecting telemetry data

In order to improve this Jenkins library we are collecting telemetry data.
Data is send using [`com.sap.piper.pushToSWA`](https://github.com/SAP/jenkins-library/blob/master/src/com/sap/piper/Utils.groovy)

Following data (non-personal) is collected for example:

* Hashed job url, e.g. `4944f745e03f5f79daf0001eec9276ce351d3035` hash calculation is done in your Jenkins server and no original values are transmitted
* Name of library step which has been executed, like e.g. `artifactSetVersion`
* Certain parameters of the executed steps, e.g. `buildTool=maven`

**We store the telemetry data for not longer than 6 months on premises of SAP SE.**

!!! note "Disable collection of telemetry data"
    If you do not want to send telemetry data you can easily deactivate this.

    This is done with either of the following two ways:

    1. General deactivation in your `.pipeline/config.yml` file by setting the configuration parameter `general -> collectTelemetryData: false` (default setting can be found in the [library defaults](https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml)).

        **Please note: this will only take effect in all steps if you run `setupCommonPipelineEnvironment` at the beginning of your pipeline**

    2. Individual deactivation per step by passing the parameter `collectTelemetryData: false`, like e.g. `setVersion script:this, collectTelemetryData: false`

## Example configuration

```yaml
general:
  gitSshKeyCredentialsId: GitHub_Test_SSH

steps:
  cloudFoundryDeploy:
    deployTool: 'cf_native'
    cloudFoundry:
      org: 'testOrg'
      space: 'testSpace'
      credentialsId: 'MY_CF_CREDENTIALSID_IN_JENKINS'
  newmanExecute:
    newmanCollection: 'myNewmanCollection.file'
    newmanEnvironment: 'myNewmanEnvironment'
    newmanGlobals: 'myNewmanGlobals'
```

## Access to configuration from custom scripts

Configuration is loaded into `commonPipelineEnvironment` during step [setupCommonPipelineEnvironment](steps/setupCommonPipelineEnvironment.md).

You can access the configuration values via `commonPipelineEnvironment.configuration` which will return you the complete configuration map.

Thus following access is for example possible (accessing `gitSshKeyCredentialsId` from `general` section):

```groovy
commonPipelineEnvironment.configuration.general.gitSshKeyCredentialsId
```

## Access to configuration in custom library steps

Within library steps the `ConfigurationHelper` object is used.

You can see its usage in all the Piper steps, for example [newmanExecute](https://github.com/SAP/jenkins-library/blob/master/vars/newmanExecute.groovy#L23).

## Custom default configuration

For projects that are composed of multiple repositories (microservices), it might be desired to provide custom default configurations.
To do that, create a YAML file which is accessible from your CI/CD environment and configure it in your project configuration.
For example, the custom default configuration can be stored in a GitHub repository and accessed via the "raw" URL:

```yaml
customDefaults: ['https://my.github.local/raw/someorg/custom-defaults/master/backend-service.yml']
general:
  ...
```

Note, the parameter `customDefaults` is required to be a list of strings and needs to be defined as a separate section of the project configuration.
In addition, the item order in the list implies the precedence, i.e., the last item of the customDefaults list has highest precedence.

It is important to ensure that the HTTP response body is proper YAML, as the pipeline will attempt to parse it.

Anonymous read access to the `custom-defaults` repository is required.

The custom default configuration is merged with the project's `.pipeline/config.yml`.
Note, the project's config takes precedence, so you can override the custom default configuration in your project's local configuration.
This might be useful to provide a default value that needs to be changed only in some projects.
An overview of the configuration hierarchy is given at the beginning of this page.

If you have different types of projects, they might require different custom default configuration.
For example, you might not require all projects to have a certain code check (like Whitesource, etc.) active.
This can be achieved by having multiple YAML files in the _custom-defaults_ repository.
Configure the URL to the respective configuration file in the projects as described above.
