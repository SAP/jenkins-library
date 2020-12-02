# The Vault ResourceRef

## Preconditions

Parameters that have a ResourceReference of type `vaultSecret` will be looked up from vault when all of the following things are true...

* The environment variables `PIPER_vaultAppRoleID` and `PIPER_vaultAppRoleSecretID` must both be set to the Vault AppRole role ID and to the Vault AppRole secret ID. See [Vault AppRole docs](https://www.vaultproject.io/docs/auth/approle)
* `vaultServerUrl` ist set in the `general` section of the configuration file.
* The parameter must not be set by the configuration file, as a CLI Parameter or an environment variable. Any parameter that has already been set won't be resolved via vault.

## Lookup

```
- name: token
        type: string
        description: "Token used to authenticate with the Sonar Server."
        scope:
          - PARAMETERS
        secret: true
        resourceRef:
          - type: vaultSecret
            paths:
            - $(vaultBasePath)/$(vaultPipelineName)/sonar
            - $(vaultBasePath)/__group/sonar
```

With the example above piper will check whether the the `token` parameter has already been set when the config was resolved. If `token` hasn't be resolved yet we will go through every item of the `paths` array, interpolate every string by using the already resolved config and then check whether there is a secret stored at the given path.

In case we find a secret we check whether it has a field (secrets in vault are **flat** json documents) that matches the parameters name (or one of the alias names), in the example above this would be `token`.
