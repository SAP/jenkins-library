# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## Build artifact metadata

**Default changed:** `createBuildArtifactsMetadata` was previously `false` and is now `true` by default. The step now writes artifact coordinates to `commonPipelineEnvironment/custom/goBuildArtifacts` on every run. Downstream steps such as OSC CTP scans consume this data.

Set `createBuildArtifactsMetadata: false` under the `golangBuild` step key in `.pipeline/config.yml` (or the equivalent step-level configuration for your orchestrator) to opt out.
