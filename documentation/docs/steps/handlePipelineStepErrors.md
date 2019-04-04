# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

none

## ${docGenParameters}

## ${docGenConfiguration}

## Example

```groovy
handlePipelineStepErrors (stepName: 'executeHealthCheck', stepParameters: parameters) {
  def url = new Utils().getMandatoryParameter(parameters, 'url', null)
  def statusCode = curl(url)
  if (statusCode != '200')
    error "Health Check failed: \${statusCode}"
}
```

## Example console output

If `echoDetails` is set to true the following information will be output to the console:

1. Step beginning: `--- Begin library step: ${stepName}.groovy ---`
1. Step end: `--- End library step: ${stepName}.groovy ---`
1. Step errors:

```log
----------------------------------------------------------
--- An error occurred in the library step: ${stepName}
----------------------------------------------------------
The following parameters were available to the step:
***
\${stepParameters}
***
The error was:
***
\${err}
***
Further information:
* Documentation of step \${stepName}: .../\${stepName}/
* Pipeline documentation: https://...
* GitHub repository for pipeline steps: https://...
----------------------------------------------------------
```
