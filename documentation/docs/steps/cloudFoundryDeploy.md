# cloudFoundryDeploy

## Description

Application will be deployed to a test or production space within Cloud Foundry.
Deployment can be done

* in a standard way
* in a zero downtime manner (using a [blue-green deployment approach](https://martinfowler.com/bliki/BlueGreenDeployment.html))

!!! note "Deployment supports multiple deployment tools"
    Currently the following are supported:

    * Standard `cf push` and [Bluemix blue-green plugin](https://github.com/bluemixgaragelondon/cf-blue-green-deploy#how-to-use)
    * [MTA CF CLI Plugin](https://github.com/cloudfoundry-incubator/multiapps-cli-plugin)

## Prerequsites

* Cloud Foundry organization, space and deployment user are available
* Credentials for deployment have been configured in Jenkins with a dedicated Id

    ![Jenkins credentials configuration](../images/cf_credentials.png)

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | yes |  |  |
| cloudFoundry | yes |  |  |
| deployTool | no | cf_native | cf_native, mtaDeployPlugin |
| deployType | no | standard | standard, blue-green |
| dockerImage | no | s4sdk/docker-cf-cli |  |
| dockerWorkspace | no | /home/piper |  |
| mtaDeployParameters |  | -f |  |
| mtaExtensionDescriptor | no | '' |  |
| mtaPath | no | '' |  |
| smokeTestScript | no | blueGreenCheckScript.sh (provided by library). <br />Can be overwritten using config property 'smokeTestScript' |  |
| smokeTestStatusCode | no | 200 |  |
| stashContent | no | []  |  |

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `cloudFoundry` defines a map containing following properties:
  * `apiEndpoint`: Cloud Foundry API endpoint (default: `https://api.cf.eu10.hana.ondemand.com`)
  * `appName`: App name of application to be deployed (optional)
  * `credentialsId`: Credentials to be used for deployment (mandatory)
  * `manifest`: Manifest to be used for deployment
  * `org`: Cloud Foundry target organization (mandatory)
  * `space`: Cloud Foundry target space (mandatory)

    Example: `cloudFoundry: [apiEndpoint: 'https://test.server.com', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', org: 'cfOrg', space: 'cfSpace']`

!!! note
    It is also possible to use following configuration parameters instead of `cloudFoundry` map:

    - cfApiEndpoint
    - cfAppName
    - cfCredentialsId
    - cfManifest
    - cfOrg
    - cfSpace

!!! note
    Due to [an incompatible change](https://github.com/cloudfoundry/cli/issues/1445) in the Cloud Foundry CLI, multiple buildpacks are not supported by this step.
    If your `application` contains a list of `buildpacks` instead a single `buildpack`, this will be automatically re-written by the step when blue-green deployment is used.

* `deployTool` defines the tool which should be used for deployment.
* `deployType` defines the type of deployment, either `standard` deployment which results in a system downtime or a zero-downtime `blue-green` deployment.
* `dockerImage` defines the Docker image containing the deployment tools (like cf cli, ...) and `dockerWorkspace` defines the home directory of the default user of the `dockerImage`
* `smokeTestScript` allows to specify a script which performs a check during blue-green deployment. The script gets the FQDN as parameter and returns `exit code 0` in case check returned `smokeTestStatusCode`. More details can be found [here](https://github.com/bluemixgaragelondon/cf-blue-green-deploy#how-to-use) <br /> Currently this option is only considered for deployTool `cf_native`.
* `stashContent` defines the stash names which should be unstashed at the beginning of the step. This makes the files available in case the step is started on an empty node.

### Deployment with cf_native

* `appName` in `cloudFoundry` map (or `cfAppName`) defines the name of the application which will be deployed to the Cloud Foundry space.
* `manifest` in `cloudFoundry` maps (or `cfManifest`) defines the manifest to be used for Cloud Foundry deployment.

!!! note
    Cloud Foundry supports the deployment of multiple applications using a single manifest file.
    This option is supported with Piper.

    In this case define `appName: ''` since the app name for the individual applications have to be defined via the manifest.
    You can find details in the [Cloud Foundry Documentation](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#multi-apps)

### Deployment with mtaDeployPlugin

* `mtaPath` define path to *.mtar for deployment.
* `mtaExtensionDescriptor` defines additional extension descriptor file for deployment.
* `mtaDeployParameters` defines additional parameters passed to mta deployment.

## Step configuration

The following parameters can also be specified as step/stage/general parameters using the [global configuration](../configuration.md):

* cloudFoundry
* deployUser
* deployTool
* deployType
* dockerImage
* dockerWorkspace
* mtaDeployParameters
* mtaExtensionDescriptor
* mtaPath
* smokeTestScript
* smokeTestStatusCode
* stashContent

## Example

```groovy
cloudFoundryDeploy(
    script: script,
    deployType: 'blue-green',
    cloudFoundry: [apiEndpoint: 'https://test.server.com', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', org: 'cfOrg', space: 'cfSpace'],
    deployTool: 'cf_native'
)
```
