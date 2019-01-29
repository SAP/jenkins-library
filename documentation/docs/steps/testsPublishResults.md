# testsPublishResults

## Description

This step can publish test results from various sources.

## Prerequsites

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

Available parameters:

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | yes | | |
| `failOnError` | no | `false` | `true`, `false` |
| junit | no | `false` | true, false |
| jacoco | no | `false` | true, false |
| cobertura | no | `false` | true, false |
| jmeter | no | `false` | true, false |

* `script` - The common script environment of the Jenkinsfile running.
    Typically the reference to the script calling the pipeline step is provided
    with the `this` parameter, as in `script: this`.
    This allows the function to access the [`commonPipelineEnvironment`](commonPipelineEnvironment.md)
    for retrieving, for example, configuration parameters.
* `failOnError` - JUnit sets the build result to `UNSTABLE` in case there are any failing tests. If `failOnError` it set the step will fail be build if the build result is set to `UNSTABLE`.
* `junit` - Publishes test results files in JUnit format with the [JUnit Plugin](https://plugins.jenkins.io/junit).
* `jacoco` - Publishes code coverage with the [JaCoCo plugin](https://plugins.jenkins.io/jacoco) .
* `cobertura` - Publishes code coverage with the [Cobertura plugin](https://plugins.jenkins.io/cobertura).
* `jmeter` - Publishes performance test results with the [Performance plugin](https://plugins.jenkins.io/performance).

Each of the parameters `junit`, `jacoco`, `cobertura` and `jmeter` can be set to `true` or `false` but also to a map of parameters to hand in different settings for the tools.

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

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/*.jtl'` |  |
| errorFailedThreshold | no | `20` |  |
| errorUnstableThreshold | no | `10` |  |
| errorUnstableResponseTimeThreshold | no | `` |  |
| relativeFailedThresholdPositive | no | `0` |  |
| relativeFailedThresholdNegative | no | `0` |  |
| relativeUnstableThresholdPositive | no | `0` |  |
| relativeUnstableThresholdNegative | no | `0` |  |
| modeOfThreshold | no | `false` | true, false |
| modeThroughput | no | `false` | true, false |
| nthBuildNumber | no | `0` |  |
| configType | no | `PRT` |  |
| failBuildIfNoResultFile | no | `false` | true, false |
| compareBuildPrevious | no | `true` | true, false |
| archive | no | `false` | true, false |
| allowEmptyResults | no | `true` | true, false |

## Step configuration

Following parameters can also be specified as step parameters using the global configuration file:

* `junit`
* `jacoco`
* `cobertura`
* `jmeter`

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
