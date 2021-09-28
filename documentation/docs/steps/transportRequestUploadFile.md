# ${docGenStepName}

## ${docGenDescription}

**Note:** This step is deprecated. Use the transport type specific steps instead.

| Type  | Step |
| ------------- | ------------- |
| SOLMAN  | [transportRequestUploadSOLMAN](transportRequestUploadSOLMAN.md)  |
| RFC  | [transportRequestUploadRFC](transportRequestUploadRFC.md)  |
| CTS  | [transportRequestUploadCTS](transportRequestUploadCTS.md)  |

You can keep most of the step specific configuration parameters in your configuration file `config.yml` untouched. The new steps support the old naming convention. However, it is recommended to adjust your parameters to the new steps.

Following parameters are not supported anymore. Adjust as indicated.

| Unsupported Parameter | Change Notice |
| ------------- | ------------- |
| changeManagement/`<type>`/docker/envVars | Use `dockerEnvVars` instead. |
| changeManagement/`<type>`/docker/image | Use `dockerImage` instead. |
| changeManagement/`<type>`/docker/options | Use `dockerOptions` instead. |
| changeManagement/`<type>`/docker/pullImage | Use `dockerPullImage` instead. |
| changeManagement/git/format | This parameter has been dropped. Make sure that your change document IDs and transport request IDs are part of the Git commit message body. |

```yaml
general:
  changeManagement:
    type: 'RFC'
# old
    rfc:
      docker:
        image: 'my/rfc-client'

#new
steps:
  transportRequestUploadRFC:
    dockerImage: 'my/rfc-client'
```

**Note:** The new steps do not comprise the retrieval of the change document ID and the transport request ID from the Git repository. Use the steps [transportRequestDocIDFromGit](transportRequestDocIDFromGit.md) and [transportRequestReqIDFromGit](transportRequestReqIDFromGit.md) instead.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

The step is configured using a customer configuration file provided as
resource in an custom shared library.

```groovy
@Library('piper-lib-os@master') _

// the shared lib containing the additional configuration
// needs to be configured in Jenkins
@Library('foo@master') __

// inside the shared lib denoted by 'foo' the additional configuration file
// needs to be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this,
                             customDefaults: 'myConfig.yml'
```

Example content of `'resources/myConfig.yml'` in branch `'master'` of the repository denoted by
`'foo'`:

```yaml
general:
  changeManagement:
    changeDocumentLabel: 'ChangeDocument\s?:'
    cmClientOpts: '-Djavax.net.ssl.trustStore=<path to truststore>'
    credentialsId: 'CM'
    type: 'SOLMAN'
    endpoint: 'https://example.org/cm'
    git:
      from: 'HEAD~1'
      to: 'HEAD'
      format: '%b'
```

The properties configured in section `'general/changeManagement'` are shared between all change management related steps.

The properties can also be configured on a per-step basis:

```yaml
  [...]
  steps:
    transportRequestUploadFile:
      applicationId: 'FOO'
      changeManagement:
        type: 'SOLMAN'
        endpoint: 'https://example.org/cm'
        [...]
```

The parameters can also be provided when the step is invoked. For examples see below.

## CTS Uploads

In order to be able to upload the application, it is required to build the application, e.g. via [npmExecuteScripts](npmExecuteScripts.md). The content of the app needs to be provided in a folder named `dist` in the root level of the project.

**Note:** Do not use the `mtaBuild` step. The MTA Build Tool `mta` is dedicated to the SAP Business Technology Platform. It does neither create the expected `dist` folder nor the compliant content.

Although the name of the step `transportRequestUploadFile` might suggest something else, in this case a folder needs
to be provided. The application, which is provided in the `dist` folder is zipped and uploaded by the fiori toolset
used for performing the upload.

For `CTS` related uploads we use a node based toolset. When running in a docker environment a standard node
image can be used. In this case the required deploy tool dependencies will be installed prior to the deploy.
It is also possible to provide a docker image which already contains the required deploy tool
dependencies (`config.changeManagement.cts.nodeDocker.image`). In this case an empty list needs to be provided
as `config.changeManagement.cts.deployToolDependencies`. Using an already pre-configured docker image speeds-up
the deployment step, but comes with the disadvantage of having
to maintain and provision the corresponding docker image.

When running in an environment without docker, it is recommanded to install the deploy tools manually on the
system and to provide an empty list for the deploy tool dependencies (`config.changeManagement.cts.deployToolDependencies`).

### Examples

#### Upload based on preconfigured image

```groovy
transportRequestUploadFile script: this,
            changeManagement: [
                credentialsId: 'CRED_ID', // credentials needs to be defined inside Jenkins
                type: 'CTS',
                endpoint: 'https://example.org:8000',
                client: '001',
                cts: [
                    nodeDocker: [
                        image: 'docker-image-name',
                        pullImage: true,  // needs to be set to false in case the image is
                                          // only available in the local docker cache (not recommended)
                    ],
                    npmInstallOpts: [],
                deployToolDependencies: [], // empty since we use an already preconfigured image
                ],
            ],
            applicationName: 'APP',
            abapPackage: 'ABABPACKAGE',
            transportRequestId: 'XXXK123456', // can be omitted when resolved via commit history
            applicationDescription: 'An optional description' // only used in case a new application is deployed
                                                              // description is not updated for re-deployments
    }

```

#### Upload based on a standard node image

```groovy
        transportRequestUploadFile script: this,
            changeManagement: [
                credentialsId: 'CRED_ID', // credentials needs to be defined inside Jenkins
                type: 'CTS',
                endpoint: 'https://example.org:8000',
                client: '001',
                cts: [
                    npmInstallOpts: [
                        '--verbose', // might be benefical for troubleshooting
                        '--registry', 'https://your.npmregistry.org/', // an own registry can be specified here
                    ],
                ],
            ],
            applicationName: 'APP',
            abapPackage: 'ABABPACKAGE',
            transportRequestId: 'XXXK123456', // can be omitted when resolved via commit history
            applicationDescription: 'An optional description' // only used in case a new application is deployed
                                                              // description is not updated for re-deployments
    }

```

## Exceptions

* `IllegalArgumentException`:
  * If the change id is not provided (`SOLMAN` only).
  * If the transport request id is not provided.
  * If the application id is not provided (`SOLMAN` only).
  * If the file path is not provided.
* `AbortException`:
  * If the upload fails.

## Example

```groovy
// SOLMAN
transportRequestUploadFile(
  script: this,
  changeDocumentId: '001',   // typically provided via git commit history
  transportRequestId: '001', // typically provided via git commit history
  applicationId: '001',
  filePath: '/path',
  changeManagement: [
    type: 'SOLMAN'
    endpoint: 'https://example.org/cm'
  ]
)
// CTS

transportRequestUploadFile(
  script: this,
  transportRequestId: '001', // typically provided via git commit history
  changeManagement: [
    type: 'CTS'
    endpoint: 'https://example.org/cm',
    client: '099',
  ],
  applicationName: 'myApp',
  abapPackage: 'MYPACKAGE',
)
```
