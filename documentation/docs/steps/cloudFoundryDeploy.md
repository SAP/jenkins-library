# ${docGenStepName}

## ${docGenDescription}

### Additional Hints

#### Standard CF deployments

`deployType` parameter defaults to value `standard`.<br>
This means that CF CLI is called by piper and command `cf push` is run by piper

#### Blue green deployments

**With CF CLI**

* Blue green deployments are deprecated, but [rolling deployment strategy](https://docs.cloudfoundry.org/devguide/deploy-apps/rolling-deploy.html) is supported.<br>
* For rolling deployment strategy , set parameter `cfNativeDeployParameters:'--strategy rolling'`
  
**With [MTA CF CLI Plugin](https://github.com/cloudfoundry-incubator/multiapps-cli-plugin) for MTA applications**

The Multiapps Plugin offers 2 different strategies:<br>

* [Blue-Green Deployment Strategy](https://github.com/SAP-samples/cf-mta-examples/tree/main/blue-green-deploy-strategy) - where the production environments are called “live” and “idle” during deployment. This strategy is activated with `mtaDeployParameters: --strategy blue-green --skip-testing-phase` and `deployType=standard`. After deployment, appnames are not appeneded by any suffix like `-live` or `-idle`.<br>
* [Legacy Blue-Green Deployment](https://github.com/SAP-samples/cf-mta-examples/tree/main/blue-green-deploy-legacy) - where the productive environments are called “blue” and “green. Activated by `deployType=blue-green`. After deployment, appnames are appeneded by suffix like `-blue` or `-green`
  
Following table summarizes the different combinations of the step parameters `deployType` and `deployTool` and their impact.
Parameter `buildTool`  is used to differentiate between MTA and Non MTA applications. If `buildTool` is not available in the environment, user will have to provide `deployTool` explicitly.

#### Deployment Strategy Comparison

This table compares deployment strategies for MTA and Non-MTA applications.

| deployType      | MTA Applications | Non MTA Applications |
|---------------|-----------------|----------------------|
| **standard**   | deployTool = mtaDeployPlugin  <br> Uses MTA plugin, <br> Command run `cf deploy` | deployTool = cf_native  <br> cf CLI used <br> Command `cf push` <br> Requires Manifest file and app name <br> appname can be provided via config or manifest file. |
| **blue-green** | deployTool = mtaDeployPlugin, <br> Uses MTA plugin <br> Command run `cf deploy bgdeploy` | Deprecated. <br> **Alternative:** Rolling deployment strategy by setting <br> `cfNativeDeployParameters = '--strategy rolling'` |
|               | **deployDockerImage not supported** | **deployDockerImage supported**<br>Docker credentials can only be provided as Jenkins environment variable. |

 !!! note

- Due to [an incompatible change](https://github.com/cloudfoundry/cli/issues/1445) in the Cloud Foundry CLI, multiple buildpacks are not supported by this step.
    If your `application` contains a list of `buildpacks` instead of a single `buildpack`, this will be automatically re-written by the step when blue-green deployment is used.
- Cloud Foundry supports the deployment of multiple applications using a single manifest file.
    This option is supported with project "Piper".
    In this case, define `appName: ''` since the app name for the individual applications has to be defined via the manifest.
    You can find details in the [Cloud Foundry Documentation](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#multi-apps)
- Recommended way to do **docker image deployments** is via [Kubernetes Piper step](https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/steps/kubernetesDeploy/).This step is not capable of deploying docker images built in the same pipeline using other piper steps[i.e., kanikoExecute].

## Prerequisites

* Cloud Foundry organization, space and deployment users are available
* Credentials for deployment have been configured in Jenkins or Vault.

## ${docGenParameters}

## ${docGenConfiguration}
