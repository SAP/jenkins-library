# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step is designed for **GitHub Actions** runners where Docker with BuildKit is natively available.
It is **not supported** on Jenkins or Azure Pipelines — use [`kanikoExecute`](kanikoExecute.md) for those orchestrators.

When pushing to a container registry, provide credentials via one of:

- A Docker `config.json` file via the `dockerConfigJSON` parameter (can be sourced from Vault using `dockerConfigFileVaultSecretName`)
- The `containerRegistryUser` and `containerRegistryPassword` parameters (credentials will be written to `~/.docker/config.json`)
- GitHub Actions secrets mapped to environment variables

## Migration from kanikoExecute

If you are currently using `kanikoExecute` on GitHub Actions, `dockerBuild` is the recommended replacement.
Kaniko is no longer actively maintained.

Most parameters are identical. Key differences:

| kanikoExecute | dockerBuild | Notes |
|---------------|-------------|-------|
| `buildOptions: ["--skip-tls-verify-pull"]` | `buildOptions: []` | Kaniko-specific flags are not needed |
| `containerPreparationCommand: "rm -f /kaniko/.docker/config.json"` | *(not needed)* | Docker BuildKit does not require preparation |
| `dockerConfigJsonCredentialsId` (Jenkins secret) | `dockerConfigJSON` via env var or Vault | GitHub Actions uses different secret mechanisms |
| `buildOptions: ["--build-arg=MY_VAR"]` | `buildOptions: ["--build-arg=MY_VAR"]` | Identical syntax |
| `buildOptions: ["--destination", "reg/img:tag"]` | `buildOptions: ["-t", "reg/img:tag"]` | Different flag name for destination |

All other parameters (`containerImage`, `containerImageName`, `containerImageTag`, `containerRegistryUrl`,
`containerMultiImageBuild`, `multipleImages`, `createBOM`, `createBuildArtifactsMetadata`, `readImageDigest`,
`registryMirrors`, `customTlsCertificateLinks`, etc.) work the same way.

## Example

### GitHub Actions workflow

```yaml
steps:
  - name: Build container image
    uses: SAP/project-piper-action@master
    with:
      command: dockerBuild
      flags: >-
        --containerImageName myImage
        --containerImageTag 1.0.0
        --containerRegistryUrl https://ghcr.io
```

### Pipeline configuration (`.pipeline/config.yml`)

```yaml
steps:
  dockerBuild:
    containerRegistryUrl: https://ghcr.io
    containerImageName: myImage
    containerImageTag: 1.0.0
```

## Passing Environment Variables to Dockerfile

If you need to pass environment variables as build arguments to your Dockerfile (e.g., for `ARG` instructions), use `buildOptions`:

```yaml
steps:
  dockerBuild:
    buildOptions:
      - --build-arg=BUILD_NUMBER
```

In your Dockerfile:

```dockerfile
ARG BUILD_NUMBER
ENV BUILD_NUMBER=${BUILD_NUMBER:-dev}
RUN echo "BUILD_NUMBER during build is: $BUILD_NUMBER"
```

On GitHub Actions, environment variables from the runner are automatically available to the Docker build process.
Unlike `kanikoExecute`, you do **not** need `dockerOptions: --env` to pass variables into a separate container.

## ${docGenParameters}

## ${docGenConfiguration}
