# toolValidate

## Description

Checks the existence and compatibility of a tool, necessary for a successful pipeline execution.
In case a violation is found, an exception is raised.

## Prerequisites

none

## Parameters

| parameter        | mandatory | default                           | possible values            |
| -----------------|-----------|-----------------------------------|----------------------------|
| `tool`           | yes       |                                   | 'java', 'mta', 'neo'       |
| `home`           | yes       |                                   |                            |

* `tool` The tool that is checked for existence and compatible version.
* `home` The location in the file system where Jenkins can access the tool.

## Step configuration

none

## Side effects

none

## Exceptions

* `IllegalArgumentException`:
  * If at least one of the parameters  `tool`, `home` is not provided.
* `AbortException`:
  * If `tool` is not supported.

## Example

```groovy
toolValidate tool: 'neo', home:'/path/to/neo-java-web-sdk'
```
