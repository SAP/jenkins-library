# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* **Snyk account** - have an account on snyk.io
* **Snyk token** - have a Snyk user token

## ${docDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## Side effects

Step uses `dockerExecute` inside.

## Exceptions

none

## Example

```groovy
snykExecute script: this, snykCredentialsId: 'mySnykToken'
```
