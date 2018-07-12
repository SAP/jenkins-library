package com.sap.piper.versioning

class MtaArtifactVersioning extends ArtifactVersioning {

    protected MtaArtifactVersioning (script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        def mtaYaml = script.readYaml file: configuration.filePath
        return mtaYaml.version
    }

    @Override
    def setVersion(version) {
        def search = "version: ${getVersion()}"
        def replacement = "version: ${version}"
        script.sh "sed -i 's/${search}/${replacement}/g' ${configuration.filePath}"
    }
}
