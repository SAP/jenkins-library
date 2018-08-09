package com.sap.piper.versioning

class DlangArtifactVersioning extends ArtifactVersioning {
    protected DlangArtifactVersioning(script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        def descriptor = script.readJSON file: configuration.filePath
        return descriptor.version
    }

    @Override
    def setVersion(version) {
        def descriptor = script.readJSON file: configuration.filePath
        descriptor.version = new String(version)
        script.writeJSON file: configuration.filePath, json: descriptor
    }
}
