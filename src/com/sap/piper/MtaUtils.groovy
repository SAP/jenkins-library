package com.sap.piper

import java.util.Map

import hudson.AbortException

class MtaUtils {

    final protected script

    protected MtaUtils(script) {
        this.script = script
    }

    def generateMtaDescriptorFromPackageJson (String srcPackageJson, String targetMtaDescriptor, String applicationName)  throws Exception{
        if (!srcPackageJson) throw new IllegalArgumentException("The parameter 'srcPackageJson' can not be null or empty.")
        if (!targetMtaDescriptor) throw new IllegalArgumentException("The parameter 'targetMtaDescriptor' can not be null or empty.")
        if (!applicationName) throw new IllegalArgumentException("The parameter 'applicationName' can not be null or empty.")

        if (!script.fileExists(srcPackageJson)) throw new AbortException("'${srcPackageJson}' does not exist.")

        def dataFromJson = script.readJSON file: srcPackageJson

        def mtaData  = script.readYaml text: script.libraryResource('template_mta.yaml')

        if(!dataFromJson.name) throw new AbortException("'name' not set in the given package.json.")
        mtaData['ID'] = dataFromJson.name

        if(!dataFromJson.version) throw new AbortException("'version' not set in the given package.json.")
        mtaData['version'] = dataFromJson.version
        mtaData['modules'][0]['parameters']['version'] = "${dataFromJson.version}-\${timestamp}"
        mtaData['modules'][0]['parameters']['name'] = applicationName

        mtaData['modules'][0]['name'] = applicationName

        script.writeYaml file: targetMtaDescriptor, data: mtaData

        if (!script.fileExists(targetMtaDescriptor)) throw new AbortException("'${targetMtaDescriptor}' has not been generated.")
    }
}
