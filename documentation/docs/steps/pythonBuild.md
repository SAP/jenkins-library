# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## Build artifact metadata

**Default changed:** `createBuildArtifactsMetadata` was previously `false` and is now `true` by default. The step now writes artifact coordinates to `commonPipelineEnvironment/custom/pythonBuildArtifacts` on every run. Downstream steps such as OSC CTP scans consume this data.

When `createBuildArtifactsMetadata` is `true` (the default), the step writes artifact coordinates (name, version, and PURL if a CycloneDX BOM was produced) to `commonPipelineEnvironment/custom/pythonBuildArtifacts` as a JSON-encoded `BuildArtifacts` object. This data is consumed by downstream steps such as OSC CTP scans.

Set `createBuildArtifactsMetadata: false` under the `pythonBuild` step key in `.pipeline/config.yml` (or the equivalent step-level configuration for your orchestrator) to opt out.
