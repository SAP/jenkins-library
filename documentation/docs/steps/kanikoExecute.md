# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

When pushing to a container registry, you need to maintain the respective credentials in your Jenkins credentials store:

Kaniko expects a Docker `config.json` file containing the credential information for registries.
An explanation on how to create it can be found at the bottom of this page.

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

## Creating a Docker `config.json` file

To create a valid `.docker/config.json` file you first need to base64 encode your username and password.
This can be done using the following command:
```shell
echo -n '<username>:<password>' | base64
```

Then create a file called config.json containing the following:
```json
{
  "auths": {
    "https://index.docker.io/v1/": {
      "auth": "userPass"
    }
  }
}
```

Replace "userPass" with the encoded username/password you created earlier.

If you need to create a config for another registry, exchange `https://index.docker.io/v1/` with the URL of the custom registry.
