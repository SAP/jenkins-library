# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

The chart must exist at `chartPath` before `helmBuild` is called. For the `publish` command, credentials for the target Helm repository are required.

### Credentials in Jenkins

Store secrets as Jenkins credentials and pass their IDs to the step:

| Credential ID parameter | Credential type | Purpose |
|---|---|---|
| `kubeConfigFileCredentialsId` | Secret file | kubeconfig for cluster authentication |
| `dockerConfigJsonCredentialsId` | Secret file | Docker `config.json` for registry access |
| `targetRepositoryCredentialsId` | Username with password | Target Helm repository authentication |
| `sourceRepositoryCredentialsId` | Username with password | Source Helm repository (dependency download) |

### Credentials from Vault

All credentials can also be resolved from HashiCorp Vault. See the parameter reference for the corresponding `*VaultSecretName` and `*VaultSecretFilePath` parameters.

## ${docJenkinsPluginDependencies}

## Usage

### Minimal call (Jenkins pipeline)

```groovy
helmBuild script: this
```

### Package and publish a chart

```yaml
# .pipeline/config.yml
steps:
  helmBuild:
    helmCommand: publish
    chartPath: helm/charts/my-app
    targetRepositoryURL: https://my-helm-registry.example.com
    targetRepositoryName: my-releases
    publish: true
```

### Lint only

```yaml
# .pipeline/config.yml
steps:
  helmBuild:
    helmCommand: lint
    chartPath: helm/charts/my-app
```

### Deploy (upgrade)

```yaml
# .pipeline/config.yml
steps:
  helmBuild:
    helmCommand: upgrade
    chartPath: helm/charts/my-app
    namespace: my-namespace
```

### Build dependencies before packaging

```yaml
# .pipeline/config.yml
steps:
  helmBuild:
    helmCommand: publish
    chartPath: helm/charts/my-app
    publish: true
    packageDependencyUpdate: true
    sourceRepositoryURL: https://my-dependency-registry.example.com
    sourceRepositoryName: dependencies
```

## Creating a Bill of Materials (BOM)

Set `createBOM: true` to generate a CycloneDX 1.4 bill of materials for the container images referenced by the chart. The BOM is written to `bom-helm.xml` and can be consumed by downstream compliance steps.

```yaml
# .pipeline/config.yml
steps:
  helmBuild:
    helmCommand: publish
    chartPath: helm/charts/my-app
    publish: true
    createBOM: true
    targetRepositoryURL: https://my-helm-registry.example.com
```

The step uses [Syft](https://github.com/anchore/syft) to generate the BOM. Pin a specific Syft release with `syftDownloadUrl` if needed.

## Signing charts before publishing

`helmBuild` can PGP-sign your Helm chart with `helm package --sign` before publishing. Two secrets are required and are resolved from Vault by default:

| Step parameter | Vault secret name parameter | Default Vault secret name | Description |
|---|---|---|---|
| `signingKey` | `signingKeyVaultSecretName` | `helm-signing` | PGP key identifier (name or email on the key) |
| `signingKeyRing` | `signingKeyRingVaultSecretName` | `helm-signing-keyring` | Path to the PGP keyring file |

```yaml
# .pipeline/config.yml
steps:
  helmBuild:
    helmCommand: publish
    chartPath: helm/charts/my-app
    publish: true
    # Override the default Vault secret names if your secrets are stored under different paths:
    signingKeyVaultSecretName: my-team/helm-signing
    signingKeyRingVaultSecretName: my-team/helm-signing-keyring
```

**Vault secret structure:**

- `signingKeyVaultSecretName` resolves via `vaultSecret`: Piper reads a field named `signingKey` from the KV secret and passes its value (the PGP key identifier) directly to `helm package --sign --key`.
- `signingKeyRingVaultSecretName` resolves via `vaultSecretFile`: Piper reads a field named `signingKeyRing` from the KV secret, writes the content to a temporary file on disk, and passes the file path to `helm package --sign --keyring`. The secret value must be the binary or ASCII-armored keyring content.

Signing is skipped when neither `signingKey` nor the Vault secret name is configured. Both `signingKey` and `signingKeyRing` must be present together — configuring only one results in an error.

## Migrating from `helmExecute`

The step was renamed from `helmExecute` to `helmBuild`. The old name is kept as an alias, so existing pipelines continue to work without any change.

To update your configuration explicitly:

```yaml
# Before
steps:
  helmExecute:
    helmCommand: publish
    chartPath: helm/charts/my-app

# After (functionally identical)
steps:
  helmBuild:
    helmCommand: publish
    chartPath: helm/charts/my-app
```

No other configuration changes are needed for the rename alone.

New parameters introduced in `helmBuild` that you may want to adopt:

| Parameter | Description | Default |
|---|---|---|
| `createBOM` | Generate a CycloneDX BOM for chart images | `false` |
| `signingKeyVaultSecretName` | Vault secret name for the PGP signing key | `helm-signing` |
| `signingKeyRingVaultSecretName` | Vault secret name for the PGP keyring file | `helm-signing-keyring` |

## ${docGenParameters}

## ${docGenConfiguration}
