# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## Build artifact metadata

`createBuildArtifactsMetadata` is a new parameter for `pythonBuild`, defaulting to `true`. When enabled, the step writes artifact coordinates (name, version, and PURL if a CycloneDX BOM was produced) to `commonPipelineEnvironment/custom/pythonBuildArtifacts` as a JSON-encoded `BuildArtifacts` object. This data is consumed by downstream steps such as OSC CTP scans.

Set `createBuildArtifactsMetadata: false` under the `pythonBuild` step key in `.pipeline/config.yml` (or the equivalent step-level configuration for your orchestrator) to opt out.
