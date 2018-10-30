package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class CfManifestUtils {
    @NonCPS
    static Map transform(Map manifest) {
        if (manifest.applications[0].buildpacks) {
            manifest['applications'].each { Map application ->
                def buildpacks = application['buildpacks']
                if (buildpacks) {
                    if (buildpacks instanceof List) {
                        if (buildpacks.size > 1) {
                            throw new RuntimeException('More than one Cloud Foundry Buildpack is not supported. Please check your manifest.yaml file.')
                        }
                        application['buildpack'] = buildpacks[0]
                        application.remove('buildpacks')
                    } else {
                        throw new RuntimeException('"buildpacks" in manifest.yaml is not a list. Please check your manifest.yaml file.')
                    }
                }
            }
        }
        return manifest
    }
}
