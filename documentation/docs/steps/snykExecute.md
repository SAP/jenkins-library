# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* **Snyk account** - have an account on snyk.io
* **Snyk token** - have a Snyk user token

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

Step uses `dockerExecute` inside.

## Exceptions

none

## Example

```groovy
snykExecute script: this, snykCredentialsId: 'mySnykToken'
```
