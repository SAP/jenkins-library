package com.sap.piper


class DownloadCacheUtils {

    static boolean isEnabled(Script script) {
        script.node('master') {
            return (Boolean.valueOf(script.env.DL_CACHE_NETWORK) && Boolean.valueOf(script.env.DL_CACHE_HOSTNAME))
        }
    }

    static String getDockerOptions(Script script) {
        script.node('master') {
            String dockerNetwork = script.env.DL_CACHE_NETWORK
            if (!dockerNetwork) {
                return ''
            }
            return " --network=$dockerNetwork"
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

        String mavenSettingsTemplate = script.libraryResource("mvn_download_cache_proxy_settings.xml")
        String mavenSettings = mavenSettingsTemplate.replace('__HOSTNAME__', hostname)

        if (!script.fileExists('.pipeline')) {
            script.sh "mkdir .pipeline"
        }

        script.writeFile file: globalSettingsFilePath, text: mavenSettings
        return globalSettingsFilePath
    }
}
