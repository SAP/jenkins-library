package com.sap.icd.jenkins.versioning

class MtaArtifactVersioning extends ArtifactVersioning {

    private String baseVersion

    protected MtaArtifactVersioning (script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        baseVersion = script.readYaml(file: configuration.filePath).version
        return baseVersion
    }

    @Override
    def setVersion(version) {
        script.sh "sed -i 's/version: ${baseVersion}/version: ${newVersion}/g' ${configuration.filePath}"
    }
}
