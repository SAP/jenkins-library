# handlePipelineStepErrors

## Description
Used by other steps to make error analysis easier. Lists parameters and other data available to the step in which the error occurs.

## Prerequisites
none

## Parameters
| parameter        | mandatory | default | possible values |
| -----------------|-----------|---------|-----------------|
| `stepParameters` | yes       |         |                 |
| `stepName`       | yes       |         |                 |
| `echoDetails`    | yes       | true    | true, false     |

* `stepParameters` - The parameters from the step to be executed. The list of parameters is then shown in the console output.
* `stepName` - The name of the step executed to be shown in the console output.
* `echoDetails` - If set to true the following will be output to the console:
    1. Step beginning: `--- BEGIN LIBRARY STEP: ${stepName}.groovy ---`
    2. Step end: `--- END LIBRARY STEP: ${stepName}.groovy ---`
    3. Step errors: 
    ```
    ----------------------------------------------------------
    --- ERROR OCCURED IN LIBRARY STEP: ${stepName}
    ----------------------------------------------------------
    FOLLOWING PARAMETERS WERE AVAILABLE TO THIS STEP:
    ***
    ${stepParameters}
    ***
    ERROR WAS:
    ***
    ${err}
    ***
    FURTHER INFORMATION:
    * Documentation of step ${stepName}: .../${stepName}/
    * Pipeline documentation: https://...
    * GitHub repository for pipeline steps: https://...
    ----------------------------------------------------------
    ```

## Step configuration
none

## Side effects
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
