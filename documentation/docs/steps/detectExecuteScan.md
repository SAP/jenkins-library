# detectExecuteScan

## Description

This step executes [Synopsys Detect](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/62423113/Synopsys+Detect) scans.
Synopsys Detect command line utlity can be used to run various scans including BlackDuck and Polaris scans. This step allows users to run BlackDuck scans by default.
Please configure your BlackDuck server Url using the serverUrl parameter and the API token of your user using the apiToken parameter for this step.


## Prerequisites

You need to store the API token for the Detect service as _'Secret text'_ credential in your Jenkins system.

!!! note "minimum plugin requirement"
    This step requires [synopsys-detect-plugin](https://github.com/jenkinsci/synopsys-detect-plugin) with at least version `2.0.0`.



## Example

```groovy
detectExecuteScan script: this, scanProperties: ['--logging.level.com.synopsys.integration=TRACE']
```

## Parameters

| name | mandatory | default | possible values |
| ---- | --------- | ------- | --------------- |
| `apiToken` | Yes |  |  |
| `codeLocation` | No |  |  |
| `detectTokenCredentialsId` | Yes |  |  |
| `dockerEnvVars` | No | `[]` |  |
| `dockerImage` | No | `openjdk:11` |  |
| `dockerName` | No | `openjdk` |  |
| `dockerOptions` | No | `[-u 0]` |  |
| `dockerWorkspace` | No | `/root` |  |
| `failOn` | No | `[BLOCKER]` | `ALL`, `BLOCKER`, `CRITICAL`, `MAJOR`, `MINOR`, `NONE` |
| `groups` | No |  |  |
| `projectName` | Yes |  |  |
| `projectVersion` | Yes |  |  |
| `scanPaths` | No | `[.]` |  |
| `scanProperties` | No | `[--blackduck.signature.scanner.memory=4096 --blackduck.timeout=6000 --blackduck.trust.cert=true --detect.report.timeout=4800 --logging.level.com.synopsys.integration=DEBUG]` |  |
| `scanners` | No | `[signature]` | `signature`, `source` |
| `script` | Yes |  |  |
| `serverUrl` | No |  |  |
| `stashContent` | No | `[buildDescriptor, checkmarx]` |  |
| `verbose` | No | `false` | `true`, `false` |

 * `apiToken`: Api token to be used for connectivity with Synopsis Detect server.
 * `codeLocation`: An override for the name Detect will use for the scan file it creates.
 * `detectTokenCredentialsId`: Jenkins 'Secret text' credentials ID containing the API token used to authenticate with the Synopsis Detect (formerly BlackDuck) Server.
 * `dockerEnvVars`: Environment variables to set in the container, e.g. [http_proxy: "proxy:8080"].
 * `dockerImage`: Name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
 * `dockerName`: Kubernetes only: Name of the container launching dockerImage. SideCar only: Name of the container in local network.
 * `dockerOptions`: Docker options to be set when starting the container.
 * `dockerWorkspace`: Kubernetes only: Specifies a dedicated user home directory for the container which will be passed as value for environment variable HOME.
 * `failOn`: Mark the current build as fail based the policy categories
 * `groups`: Users groups to be assigned for the Project
 * `projectName`: Name of the Synopsis Detect (formerly BlackDuck) project.
 * `projectVersion`: Version of the Synopsis Detect (formerly BlackDuck) project.
 * `scanPaths`: List of paths which should be scanned by the Synopsis Detect (formerly BlackDuck) scan.
 * `scanProperties`: Properties passed to the Synopsis Detect (formerly BlackDuck) scan. You can find details in the [Synopsis Detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/622846/Using+Synopsys+Detect+Properties)
 * `scanners`: List of scanners to be used for Synopsis Detect (formerly BlackDuck) scan.
 * `script`: The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` for retrieving, e.g. configuration parameters.
 * `serverUrl`: Server url to the Synopsis Detect (formerly BlackDuck) Server.
 * `stashContent`: Specific stashes that should be considered for the step execution.
 * `verbose`: verbose output


## Step Configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections of the config.yml the configuration is possible:

| parameter | general | step/stage |
| --------- | ------- | ---------- |
| `apiToken` |  | X |
| `codeLocation` |  | X |
| `dockerEnvVars` |  | X |
| `dockerImage` |  | X |
| `dockerName` |  | X |
| `dockerOptions` |  | X |
| `dockerWorkspace` |  | X |
| `failOn` |  | X |
| `groups` |  | X |
| `projectName` |  | X |
| `projectVersion` |  | X |
| `scanPaths` |  | X |
| `scanProperties` |  | X |
| `scanners` |  | X |
| `serverUrl` |  | X |
| `stashContent` |  | X |
| `verbose` | X |  |

