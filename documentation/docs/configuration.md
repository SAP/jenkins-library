# Configuration

Configuration is done via a yml-file, located at `.pipeline/config.yml` in the **master branch** of your source code repository.

Your configuration inherits from the default configuration located at [https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml](https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml).

!!! caution "Adding custom parameters"
    Please note that adding custom parameters to the configuration is at your own risk.
    We may introduce new parameters at any time which may clash with your custom parameters.

Configuration of the Piper steps as well the Piper templates can be done in a hierarchical manner.

1. Directly passed step parameters will always take precedence over other configuration values and defaults
2. Stage configuration parameters define a Jenkins pipeline stage dependent set of parameters (e.g. deployment options for the `Acceptance` stage)
3. Step configuration defines how steps behave in general (e.g. step `cloudFoundryDeploy`)
4. General configuration parameters define parameters which are available across step boundaries
5. Default configuration comes with the Piper library and is always available

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
    If you do not want to send telemetry data which helps this open source project to improve you can easily deactivate this.

    This is done with either of the following two ways:

    1. General deactivation in your `.pipeline/config.yml` file by setting the configuration parameter `general -> collectTelemetryData: false` (default setting can be found in the [library defaults](https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml)).

        **Please note: this will only take effect in all steps if you run `setupCommonPipelineEnvironment` at the beginning of your pipeline**

    2. Individual deaction per step by passing the parameter `collectTelemetryData: false`, like e.g. `setVersion script:this, collectTelemetryData: false`


## Example configuration

```
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
```
commonPipelineEnvironment.configuration.general.gitSshKeyCredentialsId
```

## Access to configuration in custom library steps

Within library steps the `ConfigurationHelper` object is used.

You can see its usage in all the Piper steps, for example [newmanExecute](https://github.com/SAP/jenkins-library/blob/master/vars/newmanExecute.groovy#L23).



