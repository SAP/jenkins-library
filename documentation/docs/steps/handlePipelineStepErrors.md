# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

none

## ${docGenParameters}

## Step configuration

none

## Exceptions

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
