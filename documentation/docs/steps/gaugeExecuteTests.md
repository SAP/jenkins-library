# ${docGenStepName}

## ${docGenDescription}

## Prerequsites

none

## ${docDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

We recommend to define values of step parameters via [config.yml file](../configuration.md).

## Example

Pipeline step:

```groovy
gaugeExecuteTests script: this, testServerUrl: 'http://test.url'
```
