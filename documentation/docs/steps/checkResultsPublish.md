# checkResultsPublish

## Description
This step can publish static check results from various sources.

## Prerequisites
* **static check result files** - To use this step, there must be static check result files available.

## Parameters
| parameter      | mandatory | default                           | possible values    |
| ---------------|-----------|-----------------------------------|--------------------|
|   |   |   |   |

* `<parameter>` - Detailed description of each parameter.

## Return value
none

## Side effects
none

## Exceptions
* `ExceptionType`
    * List of cases when exception is thrown.

## Example
```groovy
// publish java results from pmd, cpd, checkstyle & findbugs
checkResultsPublish archive: true, pmd: true, cpd: true, findbugs: true, checkstyle: true, aggregation: [thresholds: [fail: [high: 0]]]
```

```groovy
// publish javascript results from ESLint
checkResultsPublish archive: true, eslint: [pattern: '**/result-file-with-fancy-name.xml'], aggregation: [thresholds: [fail: [high: 0, normal: 10]]]
```

```groovy
// publish scala results from scalastyle
checkResultsPublish archive: true, checkstyle: [pattern: '**/target/scalastyle-result.xml']
```

```groovy
// publish python results from pylint
checkResultsPublish archive: true, pylint: [pattern: '**/target/pylint.log']
```
