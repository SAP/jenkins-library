package com.sap.piper.versioning

class GolangArtifactVersioning extends ArtifactVersioning {
    protected GolangArtifactVersioning(script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        return script.readFile(configuration.filePath).split('\n')[0].trim()
    }

    @Override
    def setVersion(version) {
        script.writeFile file: configuration.filePath, text: version
    }
}
