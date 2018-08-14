package com.sap.piper.versioning

class SbtArtifactVersioning extends ArtifactVersioning {
    protected SbtArtifactVersioning(script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        def sbtDescriptorJson = script.readJSON file: configuration.filePath
        return sbtDescriptorJson.version
    }

    @Override
    def setVersion(version) {
        def sbtDescriptorJson = script.readJSON file: configuration.filePath
        sbtDescriptorJson.version = new String(version)
        script.writeJSON file: configuration.filePath, json: sbtDescriptorJson
    }
}
