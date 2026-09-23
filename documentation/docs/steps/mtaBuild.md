# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

While using a custom docker file, ensure that the following tools are installed:

* **multitarget application archive builder 1.0.6 or compatible version** - can be downloaded from [SAP Development Tools](https://tools.hana.ondemand.com/#cloud).
* **Java 8 or compatible version** - necessary to run the *multitarget application archive builder* itself and to build Java modules.
* **NodeJS installed** - the multitarget application archive builder uses `npm` to download node module dependencies such as `grunt`.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

1. The file name of the resulting archive is written to the `commonPipelineEnvironment` with variable name `mtarFileName`.

## Exceptions

* `AbortException`:
  * If there is an invalid `buildTarget`.
  * If there is no key `ID` inside the `mta.yaml` file.

## Example

```groovy
dir('/path/to/FioriApp'){
  mtaBuild script:this, buildTarget: 'NEO'
}
def mtarFilePath = commonPipelineEnvironment.getMtarFilePath()
```

## Build artifact metadata

**Default changed:** `createBuildArtifactsMetadata` was previously `false` and is now `true` by default. The step now writes artifact coordinates to `commonPipelineEnvironment/custom/mtaBuildArtifacts` on every run. Downstream steps such as OSC CTP scans consume this data.

Set `createBuildArtifactsMetadata: false` under the `mtaBuild` step key in `.pipeline/config.yml` (or the equivalent step-level configuration for your orchestrator) to opt out.
