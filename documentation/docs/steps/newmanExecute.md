# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* prepared Postman with a test collection

## ${docGenParameters}

## ${docGenConfiguration}

## Side effects

Step uses `dockerExecute` inside.

## Exceptions

none

## Example

Pipeline step:

```groovy
newmanExecute script: this
```

This step should be used in combination with `testsPublishResults`:

```groovy
newmanExecute script: this, failOnError: false
testsPublishResults script: this, junit: [pattern: '**/newman/TEST-*.xml']
```
