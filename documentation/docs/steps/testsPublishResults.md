# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* **test result files** - To use this step, there must be test result files available.
* installed plugins:
  * [junit](https://plugins.jenkins.io/junit)
  * [jacoco](https://plugins.jenkins.io/jacoco)
  * [cobertura](https://plugins.jenkins.io/cobertura)
  * [performance](https://plugins.jenkins.io/performance)

## Pipeline configuration

none

## Explanation of pipeline step

Usage of pipeline step:

```groovy
testsPublishResults(
  junit: [updateResults: true, archive: true],
  jacoco: [archive: true]
)
```

## ${docGenParameters}

### junit

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/TEST-*.xml'` |  |
| archive | no | `false` | true, false |
| updateResults | no | `false` | true, false |
| allowEmptyResults | no | `true` | true, false |

### jacoco

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/target/*.exec'` |  |
| include | no | `''` | `'**/*.class'` |
| exclude | no | `''` | `'**/Test*'` |
| archive | no | `false` | true, false |
| allowEmptyResults | no | `true` | true, false |

### cobertura

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/target/coverage/cobertura-coverage.xml'` |  |
| archive | no | `false` | true, false |
| allowEmptyResults | no | `true` | true, false |
| onlyStableBuilds | no | `true` | true, false |

### jmeter

| parameter | mandatory | default      | possible values |
| ----------|-----------|--------------|-----------------|
| pattern | no | `'**/*.jtl'` |  |
| errorFailedThreshold | no | `20`         |  |
| errorUnstableThreshold | no | `10`         |  |
| errorUnstableResponseTimeThreshold | no | ``           |  |
| relativeFailedThresholdPositive | no | `0`          |  |
| relativeFailedThresholdNegative | no | `0`          |  |
| relativeUnstableThresholdPositive | no | `0`          |  |
| relativeUnstableThresholdNegative | no | `0`          |  |
| modeOfThreshold | no | `false`      | true, false |
| modeThroughput | no | `false`      | true, false |
| nthBuildNumber | no | `0`          |  |
| configType | no | `PRT`        |  |
| failBuildIfNoResultFile | no | `false`      | true, false |
| compareBuildPrevious | no | `true`       | true, false |
| archive | no | `false`      | true, false |
| allowEmptyResults | no | `true`       | true, false |
| filterRegex | no | ' '           |  |

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Exceptions

none

## Example

```groovy
// publish test results with coverage
testsPublishResults(
  junit: [updateResults: true, archive: true],
  jacoco: [archive: true]
)
```

```groovy
// publish test results with coverage
testsPublishResults(
  junit: [pattern: '**/target/TEST*.xml', archive: true],
  cobertura: [pattern: '**/target/coverage/cobertura-coverage.xml']
)
```
