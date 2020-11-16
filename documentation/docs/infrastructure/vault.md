# Vault for Pipeline Secrets

Project "Piper" also supports fetching your pipeline secrets directly from [Vault](https://www.hashicorp.com/products/vault).
Currently Vault's key value engine is supported in version 1 and 2, although we recommend version 2 since it supports versioning of secrets

Parameters that support being fetched from Vault are marked with the Vault Label in the Step Documentation.

![Vault Label](../images/parameter-with-vault-support.png)

## Vault Setup

The first step to store your pipeline secrets in vault, is to enable a the [Key-Value Engine](https://www.vaultproject.io/docs/secrets/kv/kv-v2). And then create a policy which grants read access to the key value engine.
For Piper to authenticate against Vault, [AppRole](https://www.vaultproject.io/docs/auth/approle) authentication must be enabled in your Vault instance.
You have to [create an AppRole Role](https://www.vaultproject.io/api-docs/auth/approle#create-update-approle) for Piper and assign it the necessary policies.

## Store Your Vault Credentials In Jenkins

Take the role ID from your Vault AppRole and create a Jenkins `Secret Text` credential. Do the same for the Vault AppRole secret ID.

![Create two jenkins secret text credentials](../images/jenkins-vault-credential.png)

## Pipline Configuration

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
