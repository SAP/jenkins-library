# ${docGenStepName}

## ${docGenDescription}

## Prerequsites

When pushing to a container registry, you need to maintain the respective credentials in your Jenkins credentials store:

Kaniko expects a Docker `config.json` file containing the credential information for registries.
You can create it like explained in the Docker Success Center in the articale about [How to generate a new auth in the config.json file](https://success.docker.com/article/generate-new-auth-in-config-json-file).

Please copy this file and upload it to your Jenkins for example<br />
via _Jenkins_ -> _Credentials_ -> _System_ -> _Global credentials (unrestricted)_ -> _Add Credentials_ ->

* Kind: _Secret file_
* File: upload your `config.json` file
* ID: specify id which you then use for the configuration of `dockerConfigJsonCredentialsId` (see below)

## ${docJenkinsPluginDependencies}

## Example

```groovy
kanikoExecute script:this
```

## ${docGenParameters}

## ${docGenConfiguration}
