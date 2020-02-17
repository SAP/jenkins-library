package com.sap.piper

import java.util.List

class PiperGoUtils implements Serializable {

    private static String additionalConfigFolder = ".pipeline/additionalConfigs"
    private static def steps
    private static Utils utils

    PiperGoUtils(def steps) {
        this.steps = steps
        this.utils = new Utils()
    }

    PiperGoUtils(def steps, Utils utils) {
        this.steps = steps
        this.utils = utils
    }

    void unstashPiperBin() {

        if (utils.unstash('piper-bin').size() > 0) return

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

    /*
     * The returned string can be used directly in the command line for retrieving the configuration via go
     */
    public String prepareConfigurations(List configs, String configCacheFolder) {

        for(def customDefault : configs) {
            steps.writeFile(file: "${additionalConfigFolder}/${customDefault}", text: steps.libraryResource(customDefault))
        }
        joinAndQuote(configs.reverse(), configCacheFolder)
    }

    /*
     * prefix is supposed to be provided without trailing slash
     */
    private static String joinAndQuote(List l, String prefix = '') {

        Iterable _l = []

        if(prefix == null) {
            prefix = ''
        }
        if(prefix.endsWith('/') || prefix.endsWith('\\'))
            throw new IllegalArgumentException("Provide prefix (${prefix}) without trailing slash")

        for(def e : l) {
            def _e = ''
            if(prefix.length() > 0) {
                _e += prefix
                _e += '/'
            }
            _e += e
            _l << '"' + _e + '"'
        }
        _l.join(' ')
    }
}
