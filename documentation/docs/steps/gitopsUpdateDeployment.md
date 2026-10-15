# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## GitHub App authentication

The default `authenticationMode: legacy` preserves username/password credentials,
including the existing `gitHttpsCredential` Vault lookup and overwrite behavior.
To use a GitHub App for a separate deployment repository, explicitly select
`githubApp` in this step or stage. App mode replaces any injected password with a
fresh installation token; an App authentication failure stops the step without
falling back to that password.

```yaml
general:
  vaultBasePath: kv/my-project
  vaultPipelineName: my-pipeline
steps:
  gitopsUpdateDeployment:
    authenticationMode: githubApp
    githubAppVaultSecretName: githubApp
    serverUrl: https://github.example.com/my-team/deployment-config
    branchName: main
    tool: kustomize
    filePath: environments/dev/kustomization.yaml
    deploymentName: registry.example.com/my-app
    maxPushAttempts: 3
```

Keep `appId` and `privateKey` together in one Vault secret. Piper checks its
configured `vaultPath`, `vaultBasePath/vaultPipelineName`, and
`vaultBasePath/GROUP-SECRETS` roots in that order. These are logical KV paths;
do not insert a KV-v2 `data/` segment. Only an absent secret causes a lookup at
the next root. A permission error or incomplete pair fails the step. App secrets
are never fetched in legacy mode.

Alternatively, bind `PIPER_githubAppId` and `PIPER_githubAppPrivateKey` as secret
environment variables. Supply both together; a complete direct pair takes
precedence over Vault. The private key must be an unencrypted RSA PEM in PKCS#1
or PKCS#8 format. Do not put private keys in pipeline YAML or command-line flags.

Install the App on the **target** repository with **Contents: read and write**.
The step discovers that repository's installation and requests access to only
that repository, with `contents: write`. The target HTTPS URL determines the API
host, independently of the source repository. GitHub.com uses `api.github.com`;
Enterprise uses the target origin plus `/api/v3`. `githubAppApiUrl` can override
the path on the same API host. Redirects are rejected. Custom TLS certificates
use `customTlsCertificateLinks` as for Git operations.

The installation token remains in memory, is masked in Piper logs, and is revoked
on completion or failure. A failed revocation logs a warning and leaves the
short-lived token to expire. Configured `username` is retained as the commit
author; if absent, App mode uses `x-access-token`.

Authentication does not enable stages or change deployment ordering. Enable the
step in the intended stage and retain that pipeline's build, promotion, and
branch gates. Confirm the chosen runtime contains the configured deployment
tool before removing custom extensions.

`maxPushAttempts` defaults to `1`. Opt into `2` or `3` to recover from concurrent
branch updates: the step waits two seconds before attempt two and five seconds
before attempt three, then clones the latest branch and reapplies the manifest
change. Only known branch-conflict push failures are retried. Authentication,
configuration, clone, and network failures are not retried; `forcePush` remains
unchanged.
