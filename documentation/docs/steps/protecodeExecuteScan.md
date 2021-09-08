# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

1. Create a Username / Password credential with the Protecode user in your Jenkins credential store
1. Look up your Group ID using REST API via `curl -u <username> "https://<protecode host>/api/groups/"`.

If the image is on a protected registry you can provide a Docker `config.json` file containing the credential information for the registry.
You can either use `docker login` (see the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/) for details) or you can create the file manually using the following script.

```shell
#!/bin/bash
auth=$(echo -n "$USER:$PASSWORD" | base64 -w0)
cat <<EOF > config.json
{
    "auths": {
        "$REGISTRY": {
            "auth": "$auth"
        }
    }
}
EOF
```

`REGISTRY` is the URL of the protected registry (Example: `https://index.docker.io/v1`).

Attention: If you reference the file in --dockerConfigJSON or upload the file to the Jenkins credential store, the file has to be named `config.json`.

## ${docGenParameters}

### Details

* The Protecode scan step is able to send a file addressed via parameter `filePath` to the backend for scanning it for known vulnerabilities.
* Alternatively an HTTP URL can be specified via `fetchUrl`. Protecode will then download the artifact from there and scan it.
* To support docker image scanning please provide `scanImage` with a docker like URL poiting to the image tag within the docker registry being used.
* To receive the result it polls until the job completes.
* Once the job has completed a PDF report is pulled from the backend and archived in the build
* Finally the scan result is being analysed for critical findings with a CVSS v3 score >= 7.0 and if such findings are detected the build is failed based on the configuration setting `failOnSevereVulnerabilities`.
* During the analysis all CVEs which are triaged are ignored and will not provoke the build to fail.

## ${docGenConfiguration}
