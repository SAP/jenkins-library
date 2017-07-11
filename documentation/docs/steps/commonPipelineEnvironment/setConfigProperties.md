# setConfigProperties

## Description
Sets the map of configuration properties stored in the `commonPipelineEnvironment` object.
Any existing `configProperties` map is overwritten.

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| `map`     | yes       |         |                 |

* `map` - the map of configuration properties to set in the `commonPipelineEnvironment` object.

## Return values

none

## Side effects

none

## Exceptions

none

## Example

```groovy
def map = [DEPLOY_HOST: 'deploy-host.com', DEPLOY_ACCOUNT: 'deploy-account']
commonPipelineEnvironment.setConfigProperties(map)
```

