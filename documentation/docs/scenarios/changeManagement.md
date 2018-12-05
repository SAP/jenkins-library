# Placeholder CM-Scenario

### Prerequisites

You have downloaded the Change Management Client 2.0.0 or a compatible version. See [Maven Central Repository](http://central.maven.org/maven2/com/sap/devops/cmclient/dist.cli/).

### Procedure

1. Check if in SAP Solution Manager, there is a change document in status "in development". See [Check for a Change Document in Status "In Development"](#check-for-a-change-document-in-status-in-development").
2. Create a transport request for a change document in SAP Solution Manager. See [Create a Transport Request](#create-a-transport-request).
3. (Optional) Upload a file to your transport request.
4. Release your transport request.


## Check for a Change Document in Status "In Development"

Check if in SAP Solution Manager, there is a change document in status "in development".


### Context

The change document ID is either retrieved from the Git commit history or provided through the parameter `changeDocumentId`.


### Mandatory Parameters

| Parameter | Description |
| --- | --- |
| `script` | The common script environment of the running Jenkinsfile. |
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

// inside the shared lib denoted by 'foo' the additional configuration file
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
The parameters can also be provided when the step is invoked. See [Examples](#Examples)

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

## Create a Transport Request

Create a transport request for a change document in SAP Solution Manager.

### Mandatory parameters

| Parameter | Description |
| --- | --- |
| `script` | The common script envoronment of the running Jenkinsfile. The reference to the script that calls the pipeline step is privided by the `this` parameter, as in `script: this`. This allows the function to access the `commonPipelineEnvironment` to retrieve configuration parameters. |
| `changeManagement/credentialsId` | 
