# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## Build artifact metadata

When `createBuildArtifactsMetadata` is `true` (the default), the step writes artifact coordinates (name, version, and PURL if a CycloneDX BOM was produced) to `commonPipelineEnvironment/custom/pythonBuildArtifacts` as a JSON-encoded `BuildArtifacts` object. This data is consumed by downstream steps such as OSC CTP scans.

Set `createBuildArtifactsMetadata: false` in your pipeline configuration to opt out.
