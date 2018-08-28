package com.sap.piper.versioning

class MavenArtifactVersioning extends ArtifactVersioning {
    protected MavenArtifactVersioning (script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        def mavenPom = script.readMavenPom (file: configuration.filePath)
        return mavenPom.getVersion().replaceAll(/-SNAPSHOT$/, "")
    }

    @Override
    def setVersion(version) {
        script.mavenExecute script: script, goals: 'versions:set', defines: "-DnewVersion=${version} -DgenerateBackupPoms=false", pomPath: configuration.filePath
    }
}
