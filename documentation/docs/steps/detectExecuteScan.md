# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

You need to store the API token for the Detect service as _'Secret text'_ credential in your Jenkins system.

!!! note "minimum plugin requirement"
    This step requires [synopsys-detect-plugin](https://github.com/jenkinsci/synopsys-detect-plugin) with at least version `2.0.0`.

## ${docJenkinsPluginDependencies}

## Example

```groovy
detectExecuteScan script: this, scanProperties: ['--logging.level.com.synopsys.integration=TRACE']
```

## ${docGenParameters}

## ${docGenConfiguration}
