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
        def mvnParameter = configuration.filePath == 'pom.xml' ? '' : "--file ${configuration.filePath}"
        script.sh "mvn versions:set -DnewVersion=${version} ${mvnParameter}"
    }
}
