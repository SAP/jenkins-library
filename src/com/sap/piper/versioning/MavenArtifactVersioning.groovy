package com.sap.piper.versioning

import com.sap.piper.Utils

class MavenArtifactVersioning extends ArtifactVersioning {
    protected MavenArtifactVersioning (script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        String pomFile = configuration.filePath
        String version = Utils.evaluateFromMavenPom(script, pomFile, 'project.version')
        return version.replaceAll(/-SNAPSHOT$/, "")
    }

    @Override
    def setVersion(version) {
        script.mavenExecute script: script, goals: 'versions:set', defines: "-DnewVersion=${version} -DgenerateBackupPoms=false", pomPath: configuration.filePath, m2Path: configuration.m2Path
    }
}
