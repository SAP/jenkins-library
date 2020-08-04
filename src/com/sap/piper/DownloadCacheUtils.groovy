package com.sap.piper


class DownloadCacheUtils {

    static Map injectDownloadCacheInParameters(Script script, Map parameters, BuildTool buildTool) {
        if (!isEnabled(script)) {
            return parameters
        }

        if (!parameters.dockerOptions) {
            parameters.dockerOptions = []
        }
        if (parameters.dockerOptions instanceof CharSequence) {
            parameters.dockerOptions = [parameters.dockerOptions]
        }

        if (!(parameters.dockerOptions instanceof List)) {
            throw new IllegalArgumentException("Unexpected type for dockerOptions. Expected was either a list or a string. Actual type was: '${parameters.dockerOptions.getClass()}'")
        }
        parameters.dockerOptions.add(getDockerOptions(script))

        if (buildTool == BuildTool.MAVEN || buildTool == BuildTool.MTA) {
            String globalSettingsFile = getGlobalMavenSettingsForDownloadCache(script)
            if (parameters.globalSettingsFile && parameters.globalSettingsFile != globalSettingsFile) {
                throw new IllegalArgumentException("You can not specify the parameter globalSettingsFile if the download cache is active")
            }

            parameters.globalSettingsFile = globalSettingsFile
        }

        if (buildTool == BuildTool.NPM || buildTool == BuildTool.MTA) {
            parameters['defaultNpmRegistry'] = getNpmRegistryUri(script)
        }

        return parameters
    }

    static String networkName() {
        return System.getenv('DL_CACHE_NETWORK')
    }

    static String hostname() {
        return System.getenv('DL_CACHE_HOSTNAME')
    }

    static boolean isEnabled(Script script) {
        if (script.env.ON_K8S) {
            return false
        }

        // Do not enable the DL-cache when a sidecar image is specified.
        // This is necessary because it is currently not possible to connect a container to multiple networks.
        // Can be removed when docker plugin supports multiple networks and jenkins-library implemented that feature
        // See also https://github.com/SAP/jenkins-library/issues/1864
        if (script.env.SIDECAR_IMAGE) {
            script.echo "Download cache disabled while running with sidecar image (${script.env.SIDECAR_IMAGE})"
            return false
        }

        return (networkName() && hostname())
    }

    static String getDockerOptions(Script script) {

        String dockerNetwork = networkName()
        if (!dockerNetwork) {
            return ''
        }
        return "--network=$dockerNetwork"
    }

    static String getGlobalMavenSettingsForDownloadCache(Script script) {
        String globalSettingsFilePath = '.pipeline/global_settings.xml'
        if (script.fileExists(globalSettingsFilePath)) {
            return globalSettingsFilePath
        }

        String hostname = hostname()

        if (!hostname) {
            return ''
        }

        String mavenSettingsTemplate = script.libraryResource("com.sap.piper/templates/mvn_download_cache_proxy_settings.xml")
        String mavenSettings = mavenSettingsTemplate.replace('__HOSTNAME__', hostname)

        if (!script.fileExists('.pipeline')) {
            script.sh "mkdir .pipeline"
        }

        script.writeFile file: globalSettingsFilePath, text: mavenSettings
        return globalSettingsFilePath
    }

    static String getNpmRegistryUri(Script script) {
        String hostname = hostname()

        if (!hostname) {
            return ''
        }
        String npmRegistry = "http://${hostname}:8081/repository/npm-proxy/"
        return npmRegistry
    }
}
