# handlePipelineStepErrors

## Description

Used by other steps to make error analysis easier. Lists parameters and other data available to the step in which the error occurs.

## Prerequisites

none

## Example

```groovy
handlePipelineStepErrors (stepName: 'executeHealthCheck', stepParameters: parameters) {
  def url = new Utils().getMandatoryParameter(parameters, 'url', null)
  def statusCode = curl(url)
  if (statusCode != '200')
    error "Health Check failed: ${statusCode}"
}
```

## ${docGenParameters}

## ${docGenConfiguration}


