# Change Request Management and Continuous Delivery in Hybrid Scenarios

Combine both ABAP and non-ABAP sources and transport them through a test landscape to a productive landscape.

### Prerequisites

You have downloaded the Change Management Client 2.0.0 or a compatible version. See [Maven Central Repository](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/).

### Procedure

1. Check if in SAP Solution Manager, there is a change document in status "in development". See [Check for a Change Document in Status "In Development" (`checkChangeInDevelopment`)](#check-for-a-change-document-in-status-in-development-checkchangeindevelopment).
2. (Optional) Create a transport request for a change document in SAP Solution Manager. See [(Optional) Create a Transport Request (`transportRequestCreate`)](#optional-create-a-transport-request-transportrequestcreate).
3. Upload a file to your transport request for a change document in SAP Solution Manager. See [Upload a File to Your Transport Request (`transportRequestUploadFile`)](#upload-a-file-to-a-transport-request-transportrequestuploadfile).
4. (Optional) Release your transport request for a change document in SAP Solution Manager. See [(Optional) Release a Transport Request (`transportRequestRelease`)](#optional-release-a-transport-request-transportrequestrelease).


## Check for a Change Document in Status "In Development" (`checkChangeInDevelopment`)

Check if in SAP Solution Manager, there is a change document in status "in development".


### Context

The change document ID is either retrieved from the Git commit history or provided through the parameter `changeDocumentId`.


### Mandatory Parameters

| Parameter | Description |
| --- | --- |
| `script` | The common script envoronment of the running Jenkinsfile. The reference to the script that calls the pipeline step is privided by the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` to retrieve configuration parameters. |
| `changeDocumentId` | The ID of the change document to transport. If you do not provide it specifically, it is retrieved from the Git commit history. |
| `changeManagement/credentialsId` | The ID of the credentials that are required to connect to SAP Solution Manager. The credentials have to be maintained on Jenkins. |
| `changeManagement/endpoint` | The address of SAP Solution Manager. |


For an overview of the optional parameters, see [Parameters](https://github.com/SarahNoack/jenkins-library/blob/master/documentation/docs/steps/checkChangeInDevelopment.md#parameters).

### Step Configuration

The step is configured by using a customer configuration file that is provided as a resource in a custom shared library.
```
@Library('piper-library-os@master') _

// the shared lib containing the additional configuration
// must be configured in Jenkins
@Library(foo@master') __

// in the shared lib denoted by 'foo', the additional configuration file
// must be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this, customDefaults: 'myConfig.yml'
```
Example content of `resources/myConfig.yaml` in branch `master` of the repository denoted by `foo`:
```
general:
  changeManagement:
    changeDocumentLabel: 'ChangeDocument\s?:'
    cmClientOpts: '-Djavax.net.ssl.trustStore=<path to truststore>'
    credentialsId: 'CM'
    endpoint: 'https://example.org/cm'
    git:
      from: 'HEAD~1'
      to: 'HEAD'
      format: '%b'
```
The properties configured in section `general/changeManagement` are shared between all steps related to change management.
You can also configure the properties on a per-step basis, for example:
```  
[...]
steps:
  checkChangeInDevelopment:
    changeManagement:
      endpoint: 'https://example.org/cm'
      [...]
    failIfStatusIsNotInDevelopment: true
```
The parameters can also be provided when the step is invoked. See [Examples](#Examples).

### Result

* If the change document is in status "in development", the return value is `true`.
* If the change document is not in status "in development", a `hudson.AbortException` is thrown.

For exceptions, see [Exceptions](https://github.com/SarahNoack/jenkins-library/blob/master/documentation/docs/steps/checkChangeInDevelopment.md#exceptions).

### Examples
* All mandatory parameters are provided through the configuration and the `changeDocumentId` is retrieved from the Git commit history:
```
checkChangeInDevelopment script:this
```
* An explicit endpoint is provided and the `changeDocumentId` is searched for starting from the previous commit (`HEAD~1`):
```
checkChangeInDevelopment script:this,
                         changeManagement: [
                           endpoint: 'https:example.org/cm',
                           git: [
                             from: 'HEAD~1'
                           ]
                         ]
```

## (Optional) Create a Transport Request (`transportRequestCreate`)

Create a transport request for a change document in SAP Solution Manager.

### Context

Depending on your workflow, this step is optional. If you already have a transport request, define it in the commit message, for example:

```
Lorem ipsum dolor sit amet, cum sucipat

    sed diam nonumy eirmod tempor invidunt ut labore
    et dolore magna aliquyam erat, sed diam voluptua.
    At vero eos et accusam et justo duo
    dolores et ea rebum nisi bene.

    TransportRequest: ZZZDK900026
```
Per default, it is expected that one of the commits between origin/master and the HEAD branch contains this transport request ID.

### Mandatory parameters

| Parameter | Description |
| --- | --- |
| `script` | The common script envoronment of the running Jenkinsfile. The reference to the script that calls the pipeline step is privided by the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` to retrieve configuration parameters. |
| `changeManagement/credentialsId` |  The ID of the credentials that are required to connect to SAP Solution Manager. |
| `changeManagement/endpoint` | The address of SAP Solution Manager. |

For an overview of the optional parameters, see [Parameters](https://github.com/SarahNoack/jenkins-library/blob/master/documentation/docs/steps/transportRequestCreate.md#parameters).

### Step configuration

The step is configured by using a customer configuration file that is provided as a resource in a custom shared library.
```
@Library('piper-library-os@master') _

// the shared lib containing the additional configuration
// must be configured in Jenkins
@Library(foo@master') __

// in the shared lib denoted by 'foo', the additional configuration file
// must be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this, customDefaults: 'myConfig.yml'
```
Example content of `resources/myConfig.yaml` in branch `master` of the repository denoted by `foo`:
```
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
The properties configured in section `general/changeManagement` are shared between all steps related to change management.
You can also configure the properties on a per-step basis, for example:
```
[...]
 steps:
   transportRequestCreate:
     changeManagement:
       type: 'SOLMAN'
       endpoint: 'https://example.org/cm'
       [...]
```
The parameters can also be provided when the step is invoked. See [Examples](#Examples).

### Result

As a result, you get the ID of the newly created transport request.

For exceptions, see **Exceptions** in [transportRequestCreate](https://sap.github.io/jenkins-library/steps/transportRequestCreate/).

### Example

```
def transportRequestId = transportRequestCreate script:this,
                                                changeDocumentId: '001,'
                                                changeManagement: [
                                                  type: 'SOLMAN'
                                                  endpoint: 'https://example.org/cm'
                                                ]
```

## Upload a File to a Transport Request (`transportRequestUploadFile`)

Upload a file to your transport request for a change document in SAP Solution Manager.

### Prerequisites

You have built your Java sources. For an example, see [mtaBuild](https://sap.github.io/jenkins-library/steps/mtaBuild/).

### Mandatory Parameters

| Parameter | Description |
| --- | --- |
| `script` | The common script envoronment of the running Jenkinsfile. The reference to the script that calls the pipeline step is privided by the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` to retrieve configuration parameters. |
| `changeDocumentId` | The ID of the change document related to the transport request that is to be released. The ID is retrieved from the Git commit history. |
| `transportRequestId` | The ID of the transport request to release. The ID is retreved from the Git commit history. |
| `applicationId` | The ID of the application. |
| `filePath` | The path of the file to upload. |
| `changeManagement/credentialsId` | The ID of the credentials that are required to connect to SAP Solution Manager. |
| `changeManagement/endpoint` | The address of SAP Solution Manager. |

For an overview of the optional parameters, see **Parameters** in [transportRequestUploadFile](https://sap.github.io/jenkins-library/steps/transportRequestUploadFile/).

### Step Configuration

The step is configured by using a customer configuration file that is provided as a resource in a custom shared library.
```
@Library('piper-library-os@master') _

// the shared lib containing the additional configuration
// must be configured in Jenkins
@Library(foo@master') __

// in the shared lib denoted by 'foo', the additional configuration file
// must be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this, customDefaults: 'myConfig.yml'
```
Example content of `resources/myConfig.yaml` in branch `master` of the repository denoted by `foo`:
```
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
  The properties configured in section `general/changeManagement` are shared between all steps related to change management.
  You can also configure the properties on a per-step basis, for example:
  ```
  [...]
 steps:
   transportRequestUploadFile:
     applicationId: 'FOO'
     changeManagement:
       type: 'SOLMAN'
       endpoint: 'https://example.org/cm'
       [...]
```
The parameters can also be provided when the step is invoked. See [Examples](#Examples).

### Result

For exceptions, see **Exceptions** in [transportRequestUploadFile](https://sap.github.io/jenkins-library/steps/transportRequestUploadFile/).

### Example

```
transportRequestUploadFile script:this,
                           changeDocumentId: '001',   // typically provided via git commit history
                           transportRequestId: '001', // typically provided via git commit history
                           applicationId: '001',
                           filePath: '/path',
                           changeManagement:[
                             type: 'SOLMAN'
                             endpoint: 'https://example.org/cm'
                           ]
```

## (Optional) Release a Transport Request (`transportRequestRelease`)

Release your transport request for a change document in SAP Solution Manager.

### Mandatory parameters

| Parameter | Description |
| --- | --- |
| `script` | The common script envoronment of the running Jenkinsfile. The reference to the script that calls the pipeline step is privided by the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` to retrieve configuration parameters. |
| `changeDocumentId` |  The ID of the change document related to the transport request that is to be released. The ID is retrieved from the Git commit history. |
| `transportRequestId` | The ID of the transport request to release. The ID is retreved from the Git commit history. |
| `changeManagement/credentialsId` | The ID of the credentials that are required to connect to SAP Solution Manager. |
|`changeManagement/endpoint` | The address of SAP Solution Manager. |

For an overview of the optional parameters, see **Parameters** in [transportRequestRelease](https://sap.github.io/jenkins-library/steps/transportRequestRelease/).

### Step Configuration

The step is configured by using a customer configuration file that is provided as a resource in a custom shared library.
```
@Library('piper-library-os@master') _

// the shared lib containing the additional configuration
// must be configured in Jenkins
@Library(foo@master') __

// in the shared lib denoted by 'foo', the additional configuration file
// must be located under 'resources' ('resoures/myConfig.yml')
prepareDefaultValues script: this, customDefaults: 'myConfig.yml'
```
Example content of `resources/myConfig.yaml` in branch `master` of the repository denoted by `foo`:
```
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
The properties configured in section `general/changeManagement` are shared between all steps related to change management.
You can also configure the properties on a per-step basis, for example:
```
[...]
 steps:
   transportRequestRelease:
     changeManagement:
       type: 'SOLMAN'
       endpoint: 'https://example.org/cm'
       [...]
```
The parameters can also be provided when the step is invoked. See [Examples](#Examples).

### Result

For exceptions, see **Exceptions** in [transportRequestRelease](https://sap.github.io/jenkins-library/steps/transportRequestRelease/).

### Example

```
transportRequestRelease script:this,
                        changeDocumentId: '001',
                        transportRequestId: '001',
                        changeManagement: [
                          type: 'SOLMAN'
                          endpoint: 'https://example.org/cm'
                        ]
```
