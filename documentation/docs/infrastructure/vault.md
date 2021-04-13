# Vault for Pipeline Secrets

Project "Piper" supports fetching your pipeline secrets directly from [Vault](https://www.hashicorp.com/products/vault).
Currently, Vault's key value engine is supported in version 1 and 2, although we recommend version 2 since it supports
the versioning of secrets

Parameters that support being fetched from Vault are marked with the Vault Label in the Step Documentation.

![Vault Label](../images/parameter-with-vault-support.png)

## Authenticating Piper to Vault

Piper currently supports Vault's `AppRole` and `Token` authentication. However, `AppRole` authentication is recommended
since Piper is able to regularly rotate the SecretID, which is not possible with a Token.

### AppRole Authentication

To authenticate against Vault, using [AppRole](https://www.vaultproject.io/docs/auth/approle) authentication you need to
do the following things

- Enable AppRole authentication in your vault instance.
- After that you have
  to [create an AppRole Role](https://www.vaultproject.io/api-docs/auth/approle#create-update-approle) for Piper
- Assign the necessary policies to your newly created AppRole.
- Take the **AppRole ID** and create a Jenkins `Secret Text` credential.
- Take the **AppRole Secret ID** and create a Jenkins `Secret Text` credential.

![Create two jenkins secret text credentials](../images/jenkins-vault-credential.png)

### Token Authentication

First step to use Token authentication is
to [Create a vault Token](https://www.vaultproject.io/api/auth/token#create-token)
In order to use a Vault Token for authentication you need to store the vault token inside your Jenkins instance as shown
below.

![Create a Jenkins secret text credential](../images/jenkins-vault-token-credential.png)

## Setup a Secret Store in Vault

The first step to store your pipeline secrets in Vault, is to enable a the
[Key-Value Engine](https://www.vaultproject.io/docs/secrets/kv/kv-v2). Then create a policy which grants read access to
the key value engine.

![Enable a new secret engine in vault](../images/vault-secret-engine-enable.png)

## Pipeline Configuration

For pipelines to actually use the secrets stored in Vault you need to adjust your `config.yml`

```yml
general:
  ...
  vaultAppRoleTokenCredentialsId: '<JENKINS_CREDENTIAL_ID_FOR_VAULT_APPROLE_ROLE_ID>'
  vaultAppRoleSecretTokenCredentialsId: 'JENKINS_CREDENTIAL_ID_FOR_VAULT_APPROLE_SECRET_ID'
  vaultPath: 'kv/my-pipeline' # the path under which your jenkins secrets are stored
  vaultServerUrl: '<YOUR_VAULT_SERVER_URL>'
  vaultNamespace: '<YOUR_NAMESPACE_NAME>' # if you are not using vault's namespace feature you can remove this line
  ...
```

Or if you chose to use Vault's token authentication then your  `config.yml` should look something like this.

```yaml
general:
...
vaultTokenCredentialsId: '<JENKINS_CREDENTIAL_ID_FOR_YOUR_VAULT_TOKEN>'
vaultPath: 'kv/my-pipeline' # the path under which your jenkins secrets are stored
vaultServerUrl: '<YOUR_VAULT_SERVER_URL>'
vaultNamespace: '<YOUR_NAMESPACE_NAME>' # if you are not using vault's namespace feature you can remove this line
...
```

## Configuring the Secret Lookup

When Piper is configured to lookup secrets in Vault, there are some aspects that need to be considered.

### Overwriting of Parameters

Whenever a parameter is provided via `config.yml` or passed to the CLI it gets overwritten when a secret is found in
Vault. To disable overriding parameters put a `vaultDisableOverwrite: false` on `Step` `Stage` or `General` Section in
your config.

```yaml
general:
  ...
  vaultDisableOverwrite: true
  ...
steps:
  executeBuild:
    vaultDisableOverwrite: false
    ...
```

### Skipping Vault Secret Lookup

It is also possible to skip Vault for `Steps`, `Stages` or in `General` by using the `skipVault` config parameter as
shown below.

```yaml
...
steps:
  executeBuild:
    skipVault: true   # Skip Vault Secret Lookup for this step
```

## Using vault for test credentials

Vault can be used with piper to fetch any credentials, e.g. when they need to be appended to test command. The configuration for vault test credentials can be added to **any** piper golang-based step. The configuration has to be done as follows:

```yaml
general:
  < your vault configuration > # see above
...
steps:
  < piper go step >:
    vaultTestCredentialPath: 'myTestStepCrecetials'
    vaultTestCredentialKeys: ['myAppId', 'myAppSecret']
```

The `vaultTestCredentialPath` parameter is the endpoint of your credential path in vault. Depending on your _general_ config, the lookup for the credential IDs will be done in the following order respectively locations. The first path with found test credentials will be used.

1. `<vaultPath>/<vaultTestCredentialPath>`
2. `<vaultBasePath>/<vaultPipelineName>/<vaultTestCredentialPath>`
3. `<vaultBasePath>/GROUP-SECRETS/<vaultTestCredentialPath>`

The `vaultTestCredentialKeys`parameter is a list of credential IDs. The secret value of the credential will be exposed as an environment variable prefixed by "PIPER_TESTCREDENTIAL_" and transformed to a valid variable name. For a credential ID named `myAppId` the forwarded environment variable to the step will be `PIPER_TESTCREDENTIAL_MYAPPID` containing the secret. Hyphens will be replaced by underscores and other non-alphanumeric characters will be removed.

Extended logging for vault secret fetching (e.g. found credentials and environment variable names) can be activated via `verbose: true` configuration.
