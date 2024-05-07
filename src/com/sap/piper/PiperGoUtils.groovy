package com.sap.piper

import hudson.AbortException

class PiperGoUtils implements Serializable {

    private static piperExecutable = 'piper'

    private static Script steps
    private static Utils utils

    PiperGoUtils(Script steps) {
        this.steps = steps
        this.utils = new Utils()
    }

    PiperGoUtils(Script steps, Utils utils) {
        this.steps = steps
        this.utils = utils
    }

    void unstashPiperBin() {
        // Check if the piper binary is already present
        if (steps.sh(script: "[ -x ./${piperExecutable} ]", returnStatus: true) == 0) {
            steps.echo "Found ${piperExecutable} binary in the workspace - skipping unstash"
            return
        }

        if (utils.unstash('piper-bin').size() > 0) return

        if (steps.env.REPOSITORY_UNDER_TEST && steps.env.LIBRARY_VERSION_UNDER_TEST) {
            steps.echo("Running in a consumer test, building unit-under-test binary for verification.")
            steps.dockerExecute(script: steps, dockerImage: 'golang:1.21', dockerOptions: '-u 0', dockerEnvVars: [
                REPOSITORY_UNDER_TEST: steps.env.REPOSITORY_UNDER_TEST,
                LIBRARY_VERSION_UNDER_TEST: steps.env.LIBRARY_VERSION_UNDER_TEST
            ]) {
                def piperTar = 'piper-go.tar.gz'
                def piperTmp = 'piper-tmp'
                steps.sh "wget --output-document ${piperTar} https://github.com/\${REPOSITORY_UNDER_TEST}/archive/\$LIBRARY_VERSION_UNDER_TEST.tar.gz"
                steps.sh "PIPER_TMP=${piperTmp}; rm -rf \${PIPER_TMP} && mkdir -p \${PIPER_TMP} && tar --strip-components=1 -C \${PIPER_TMP} -xf ${piperTar}"
                steps.dir(piperTmp) {
                    steps.sh "CGO_ENABLED=0 go build -tags release -ldflags \"-X github.com/SAP/jenkins-library/cmd.GitCommit=${steps.env.LIBRARY_VERSION_UNDER_TEST}\" -o ../piper . && chmod +x ../piper && chown 1000:999 ../piper"
                }
                steps.sh "rm -rf ${piperTar} ${piperTmp}"
            }
        } else {
            def libraries = getLibrariesInfo()
            String version
            libraries.each {lib ->
                if (lib.name == 'piper-lib-os') {
                    version = lib.version
                }
            }

            def fallbackUrl = 'https://github.com/SAP/jenkins-library/releases/latest/download/piper_master'
            def piperBinUrl = (version == 'master') ? fallbackUrl : "https://github.com/SAP/jenkins-library/releases/download/${version}/piper"

            boolean downloaded = downloadGoBinary(piperBinUrl)
            if (!downloaded) {
                //Inform that no Piper binary is available for used library branch
                steps.echo ("Not able to download go binary of Piper for version ${version}")
                //Fallback to master version & throw error in case this fails
                steps.retry(12) {
                    if (!downloadGoBinary(fallbackUrl)) {
                        steps.sleep(10)
                        steps.error("Download of Piper go binary failed.")
                    }
                }
            }
        }
        try {
            def piperVersion = steps.sh returnStdout: true, script: "./${piperExecutable} version"
            steps.echo "Piper go binary version: ${piperVersion}"
        } catch(AbortException ex) {
            steps.error "Cannot get piper go binary version: ${ex}"
        }
        utils.stashWithMessage('piper-bin', 'failed to stash piper binary', piperExecutable)
    }

    List getLibrariesInfo() {
        return new JenkinsUtils().getLibrariesInfo()
    }

    private boolean downloadGoBinary(url) {

        try {
            def httpStatus = steps.sh(returnStdout: true, script: "curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out '%{http_code}' --output ${piperExecutable} '${url}'")

            if (httpStatus == '200') {
                steps.sh(script: "chmod +x ${piperExecutable}")
                return true
            }
        } catch(err) {
            //nothing to do since error should just result in downloaded=false
            steps.echo "Failed downloading Piper go binary with error '${err}'. " +
                "If curl is missing, please ensure that curl is available in the Jenkins master and the agents. It is a prerequisite to run piper."
        }
        return false
    }
}
