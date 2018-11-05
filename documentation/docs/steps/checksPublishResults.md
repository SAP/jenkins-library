# checksPublishResults

## Description
This step can publish static check results from various sources.

## Prerequisites
* **static check result files** - To use this step, there must be static check result files available.
* installed plugins:
  * [pmd](https://plugins.jenkins.io/pmd)
  * [dry](https://plugins.jenkins.io/dry)
  * [findbugs](https://plugins.jenkins.io/findbugs)
  * [checkstyle](https://plugins.jenkins.io/checkstyle)
  * [warnings](https://plugins.jenkins.io/warnings)
  * [core](https://plugins.jenkins.io/core)

## Parameters
| parameter      | mandatory | default                           | possible values    |
| ---------------|-----------|-----------------------------------|--------------------|
| script | yes | | |
| aggregation | no | `true` | see below |
| tasks | no | `false` | see below |
| pmd | no | `false` | see below |
| cpd | no | `false` | see below |
| findbugs | no | `false` | see below |
| checkstyle | no | `false` | see below |
| eslint | no | `false` | see below |
| pylint | no | `false` | see below |
| archive | no | `false` | `true`, `false` |

* `aggregation` - Publishes .
* `tasks` - Searches and publishes TODOs in files with the [Task Scanner Plugin](https://wiki.jenkins-ci.org/display/JENKINS/Task+Scanner+Plugin).
* `pmd` - Publishes PMD findings with the [PMD plugin](https://plugins.jenkins.io/pmd) .
* `cpd` - Publishes CPD findings with the [DRY plugin](https://plugins.jenkins.io/dry).
* `findbugs` - Publishes Findbugs findings with the [Findbugs plugin](https://plugins.jenkins.io/findbugs).
* `checkstyle` - Publishes Checkstyle findings with the [Checkstyle plugin](https://plugins.jenkins.io/checkstyle).
* `eslint` - Publishes ESLint findings (in [JSLint format](https://eslint.org/docs/user-guide/formatters/)) with the [Warnings plugin](https://plugins.jenkins.io/warnings).
* `pylint` - Publishes PyLint findings with the [Warnings plugin](https://plugins.jenkins.io/warnings), pylint needs to run with `--output-format=parseable` option.

Each of the parameters `aggregation`, `tasks`, `pmd`, `cpd`, `findbugs`, `checkstyle`, `eslint` and `pylint` can be set to `true` or `false` but also to a map of parameters to hand in different settings for the tools.

**aggregation**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| thresholds | no | none | see [thresholds](#thresholds) |

**tasks**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/*.java'` |  |
| archive | no | `true` | `true`, `false` |
| high | no | `'FIXME'` |  |
| normal | no | `'TODO,REVISE,XXX'` |  |
| low | no |  |  |
| thresholds | no | none | see [thresholds](#thresholds) |

**pmd**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/target/pmd.xml'` |  |
| archive | no | `true` | `true`, `false` |
| thresholds | no | none | see [thresholds](#thresholds) |

**cpd**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/target/cpd.xml'` |  |
| archive | no | `true` | `true`, `false` |
| thresholds | no | none | see [thresholds](#thresholds) |

**findbugs**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/target/findbugsXml.xml, **/target/findbugs.xml'` |  |
| archive | no | `true` | true, false |
| thresholds | no | none | see [thresholds](#thresholds) |

**checkstyle**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/target/checkstyle-result.xml'` |  |
| archive | no | `true` | `true`, `false` |
| thresholds | no | none | see [thresholds](#thresholds) |

**eslint**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/eslint.jslint.xml'` |  |
| archive | no | `true` | `true`, `false` |
| thresholds | no | none | see [thresholds](#thresholds) |

**pylint**

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| pattern | no | `'**/pylint.log'` |  |
| archive | no | `true` | `true`, `false` |
| thresholds | no | none | see [thresholds](#thresholds) |

## Step configuration
Following parameters can also be specified as step parameters using the global configuration file:

* `aggregation`
* `tasks`
* `pmd`
* `cpd`
* `findbugs`
* `checkstyle`
* `eslint`
* `pylint`
* `archive`

### Thresholds

It is possible to define thresholds to fail the build on a certain count of findings. To achive this, just define your thresholds a followed for the specific check tool:

```groovy
thresholds: [fail: [all: 999, low: 99, normal: 9, high: 0]]
```

This way, the jenkins will fail the build on 1 high issue, 10 normal issues, 100 low issues or a total issue count of 1000.

The `thresholds` parameter can be set for `aggregation`, `tasks`, `pmd`, `cpd`, `findbugs`, `checkstyle`, `eslint` and `pylint`.

```groovy
checksPublishResults(
    tasks: true,
    pmd: [pattern: '**/target/pmd-results.xml', thresholds: [fail: [low: 100]]],
    cpd: [archive: false],
    aggregation: [thresholds: [fail: [high: 0]]],
    archive: true
)
```

![StaticChecks Thresholds](../images/StaticChecks_Threshold.png)

## Return value
none

## Side effects
If both ESLint and PyLint results are published, they are not correctly aggregated in the aggregator plugin.

## Exceptions
none

## Example
```groovy
// publish java results from pmd, cpd, checkstyle & findbugs
checksPublishResults archive: true, pmd: true, cpd: true, findbugs: true, checkstyle: true, aggregation: [thresholds: [fail: [high: 0]]]
```

```groovy
// publish javascript results from ESLint
checksPublishResults archive: true, eslint: [pattern: '**/result-file-with-fancy-name.xml'], aggregation: [thresholds: [fail: [high: 0, normal: 10]]]
```

```groovy
// publish scala results from scalastyle
checksPublishResults archive: true, checkstyle: [pattern: '**/target/scalastyle-result.xml']
```

```groovy
// publish python results from pylint
checksPublishResults archive: true, pylint: [pattern: '**/target/pylint.log']
```
