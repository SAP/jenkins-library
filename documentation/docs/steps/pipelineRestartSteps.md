# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

none

## ${docGenParameters}

## ${docGenConfiguration}

## Example

Usage of pipeline step:

```groovy
pipelineRestartSteps (script: this) {
  node {
    //your steps ...
  }
}
```

!!! caution
    Use `node` inside the step. If a `node` exists outside the step context, the `input` step which is triggered in the process will block a Jenkins executor.

    In case you cannot use `node` inside this step, please choose the parameter `timeoutInSeconds` carefully!

## Side effects

none

## ${docJenkinsPluginDependencies}

## Exceptions

none
