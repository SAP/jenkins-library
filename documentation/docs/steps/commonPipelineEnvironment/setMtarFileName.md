# setMtarFileName

## Description
Sets the file name of the mtar archive. The mtar archive is created in the [mtaBuild](../../steps/mtaBuild) step.
This does not change the file name of the actual mtar archive file.

## Parameters

| parameter      | mandatory | default | possible values |
| ---------------|-----------|---------|-----------------|
| `mtarFileName` | yes       |         |                 |

* `mtarFileName` - the String to be set as value for `mtarFileName` in `commonPipelineEnvironment`.

## Return values

none

## Side effects

none

## Exceptions

none

## Example

```groovy
commonPipelineEnvironment.setMtarFileName('my.file.name.mtar')
```
