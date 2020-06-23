# ${docGenStepName}

## ${docGenDescription}

## Prerequisites
1. Request creation of a team for your development group as described [here](http://go.sap.corp/protecode) and in addition request creation of a technical Protecode user through OS3 team
2. Create a Username / Password credential with the Protecode technical user in your Jenkins credential store
3. Supply the credential ID either via config.yml or on the step via parameter `protecodeCredentialsId`
4. Supply the **group ID** of the Protecode group via parameter `protecodeGroup`. You can either inquire this value from OS3 upon creation of the group or look it up yourself via REST API using `curl -u <place your user here> "https://protecode.mo.sap.corp/api/groups/"`.

## Example

Usage of pipeline step:

Workspace based:
```groovy
executeProtecodeScan script: this, filePath: 'dockerImage.tar'
```

Fetch URL:
```groovy
executeProtecodeScan script: this, fetchUrl: 'https://nexusrel.wdf.sap.corp:8443/nexus/service/local/repositories/build.releases.3rd-party.proxy.2018.04.13/content/org/alfresco/surf/spring-cmis-framework/6.11/spring-cmis-framework-6.11.jar'
```

Docker image:
```groovy
executeProtecodeScan script: this, dockerRegistryUrl: 'https://docker.wdf.sap.corp:50000', dockerImage: 'piper/yeoman:1.0-20180321110554'
```

## ${docGenParameters}

### Details:

* The Protecode scan step is able to send a file addressed via parameter `filePath` to the backend for scanning it for known vulnerabilities.
* Alternatively an HTTP URL can be specified via `fetchUrl`. Protecode will then download the artifact from there and scan it.
* To support docker image scanning please provide `dockerImage` with a docker like URL poiting to the image tag within the docker registry being used. Our step uses [skopeo](https://github.com/containers/skopeo) to download the image and sends it to Protecode for scanning.
* To receive the result it polls until the job completes.
* Once the job has completed a PDF report is pulled from the backend and archived in the build
* Finally the scan result is being analysed for critical findings with a CVSS v3 score >= 7.0 and if such findings are detected the build is failed based on the configuration setting `protecodeFailOnSevereVulnerabilities`.
* During the analysis all CVEs which are either triaged in the Protecode backend or which are excluded via configuration parameter `protecodeExcludeCVEs` are ignored and will not provoke the build to fail.

### FAQs:

* In case of `dockerImage` and the step still tries to pull and save it via docker daemon, please make sure your JaaS environment has the variable `ON_K8S` declared and set to `true`.

## ${docGenConfiguration}
