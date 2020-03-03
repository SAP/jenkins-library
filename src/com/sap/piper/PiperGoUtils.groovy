package com.sap.piper

class PiperGoUtils implements Serializable {

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

        if (utils.unstash('piper-bin').size() > 0) return

        if (steps.env.REPOSITORY_UNDER_TEST && steps.env.LIBRARY_VERSION_UNDER_TEST) {
            steps.echo("Running in a consumer test, building unit-under-test binary for verification.")
            steps.dockerExecute(script: steps, dockerImage: 'golang:1.13', dockerOptions: '-u 0', dockerEnvVars: [
                REPOSITORY_UNDER_TEST: steps.env.REPOSITORY_UNDER_TEST,
                LIBRARY_VERSION_UNDER_TEST: steps.env.LIBRARY_VERSION_UNDER_TEST
            ]) {
                steps.sh 'wget https://github.com/$REPOSITORY_UNDER_TEST/archive/$LIBRARY_VERSION_UNDER_TEST.tar.gz'
                steps.sh 'tar xzf $LIBRARY_VERSION_UNDER_TEST.tar.gz'
                steps.dir("jenkins-library-${steps.env.LIBRARY_VERSION_UNDER_TEST}") {
                    steps.sh 'CGO_ENABLED=0 go build -tags release -o ../piper . && chmod +x ../piper && chown 1000:999 ../piper'
                }
                steps.sh 'rm -rf $LIBRARY_VERSION_UNDER_TEST.tar.gz jenkins-library-$LIBRARY_VERSION_UNDER_TEST'
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
                steps.retry(5) {
                    if (!downloadGoBinary(fallbackUrl)) {
                        steps.sleep(2)
                        steps.error("Download of Piper go binary failed.")
                    }
                }

            }
        }

        utils.stashWithMessage('piper-bin', 'failed to stash piper binary', 'piper')
    }

    List getLibrariesInfo() {
        return new JenkinsUtils().getLibrariesInfo()
    }

    private boolean downloadGoBinary(url) {

        try {
            def httpStatus = steps.sh(returnStdout: true, script: "curl --insecure --silent --location --write-out '%{http_code}' --output ./piper '${url}'")

            if (httpStatus == '200') {
                steps.sh(script: 'chmod +x ./piper')
                return true
            }
        } catch(err) {
            //nothing to do since error should just result in downloaded=false
            steps.echo "Failed downloading Piper go binary with error '${err}'"
        }
        return false
    }
}
