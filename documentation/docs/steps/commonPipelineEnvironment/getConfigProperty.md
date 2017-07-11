# getConfigProperty

## Description
Gets a specific value from the configuration properties stored in the `commonPipelineEnvironment` object.

## Parameters

| parameter  | mandatory | default | possible values |
| -----------|-----------|---------|-----------------|
| `property` | yes       |         |                 |

* `property` - the specific property to be retrieved from the configuration properties stored in the `commonPipelineEnvironment` object.

## Return values

The value of the property in the configuration properties stored in the `commonPipelineEnvironment` object.

## Side effects

none

## Exceptions

none

## Example

```groovy
def deployHost = commonPipelineEnvironment.getConfigProperty('DEPLOY_HOST')
```