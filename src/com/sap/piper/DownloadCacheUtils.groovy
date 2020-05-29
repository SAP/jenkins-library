package com.sap.piper


class DownloadCacheUtils {

    static Map injectDownloadCacheInParameters(Script script, Map parameters, BuildTool buildTool) {
        if (DownloadCacheUtils.isEnabled(script)) {

            if (!parameters.dockerOptions) {
                parameters.dockerOptions = []
            }
            if (parameters.dockerOptions instanceof CharSequence) {
                parameters.dockerOptions = [parameters.dockerOptions]
            }

            if (!(parameters.dockerOptions instanceof List)) {
                throw new IllegalArgumentException("Unexpected type for dockerOptions. Expected was either a list or a string. Actual type was: '${parameters.dockerOptions.getClass()}'")
            }
            parameters.dockerOptions.add(DownloadCacheUtils.getDockerOptions(script))

            if (buildTool == BuildTool.MAVEN || buildTool == BuildTool.MTA) {
                if (parameters.globalSettingsFile) {
                    throw new IllegalArgumentException("You can not specify the parameter globalSettingsFile if the download cache is active")
                }

                parameters.globalSettingsFile = DownloadCacheUtils.getGlobalMavenSettingsForDownloadCache(script)
            }

            if (buildTool == BuildTool.NPM || buildTool == buildTool.MTA) {
                parameters['defaultNpmRegistry'] = DownloadCacheUtils.getNpmRegistryUri(script)
            }
        }

        return parameters
    }

    static boolean isEnabled(Script script) {
        if (script.env.ON_K8S) {
            return false
        }
        script.node('master') {
            String network = script.env.DL_CACHE_NETWORK
            String host = script.env.DL_CACHE_HOSTNAME
            return (network.asBoolean() && host.asBoolean())
        }
    }

    static String getDockerOptions(Script script) {
        script.node('master') {
            String dockerNetwork = script.env.DL_CACHE_NETWORK
            if (!dockerNetwork) {
                return ''
            }
            return "--network=$dockerNetwork"
        }
    }

    static String getGlobalMavenSettingsForDownloadCache(Script script) {
        String globalSettingsFilePath = '.pipeline/global_settings.xml'
        if (script.fileExists(globalSettingsFilePath)) {
            return globalSettingsFilePath
        }

        String hostname = ''
        script.node('master') {
            hostname = script.env.DL_CACHE_HOSTNAME // set by cx-server
        }

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
        String npmRegistry = ''
        script.node('master') {
            npmRegistry = "http://${script.env.DL_CACHE_HOSTNAME}:8081/repository/npm-proxy/"
        }
        return npmRegistry
    }
}
