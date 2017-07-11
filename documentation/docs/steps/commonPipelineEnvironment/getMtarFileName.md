# getMtarFileName

## Description
Gets the file name of the mtar archive. The mtar archive is created in the [mtaBuild](../../steps/mtaBuild) step.

## Parameters

none

## Return values

The mtar archive file name stored in the `commonPipelineEnvironment` object.

## Side effects

none

## Exceptions

none

## Example

```groovy
def mtarFileName = commonPipelineEnvironment.getMtarFileName()
```
