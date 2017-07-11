# setConfigProperty

## Description
Sets a specific property of the configuration stored in the `commonPipelineEnvironment` object.
Any existing property is overwritten.

## Parameters

| parameter  | mandatory | default | possible values |
| -----------|-----------|---------|-----------------|
| `property` | yes       |         |                 |
| `value`    | yes       |         |                 |

* `property` - property key to set in the `commonPipelineEnvironment` object.
* `value`- the value to set the property to.

## Return values

none

## Side effects

none

## Exceptions

none

## Example

```groovy
commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'my-deploy-host.com')
```

