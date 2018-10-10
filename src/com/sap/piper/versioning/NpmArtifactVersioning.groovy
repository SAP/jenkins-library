package com.sap.piper.versioning

class NpmArtifactVersioning extends ArtifactVersioning {
    protected NpmArtifactVersioning(script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        def packageJson = script.readJSON file: configuration.filePath
        return packageJson.version
    }

    @Override
    def setVersion(version) {
        def packageJson = script.readJSON file: configuration.filePath
        packageJson.version = new String(version)
        script.writeJSON file: configuration.filePath, json: packageJson
    }
}
