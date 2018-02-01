package com.sap.piper.versioning

class DockerArtifactVersioning extends ArtifactVersioning {
    protected DockerArtifactVersioning(script, configuration) {
        super(script, configuration)
    }

    @Override
    def getVersion() {
        if (configuration.dockerVersionSource == 'FROM')
            return  getVersionFromDockerBaseImageTag(configuration.filePath)
        else
            //standard assumption: version is assigned to an env variable
            return getVersionFromDockerEnvVariable(configuration.filePath, configuration.dockerVersionSource)
    }

    @Override
    def setVersion(version) {
        def dockerVersionDir = (configuration.dockerVersionDir?dockerVersionDir:'')
        script.dir(dockerVersionDir) {
            script.writeFile file:'VERSION', text: version
        }
    }

    def getVersionFromDockerEnvVariable(filePath, name) {
        def lines = script.readFile(filePath).split('\n')
        def version = ''
        for (def i = 0; i < lines.size(); i++) {
            if (lines[i].startsWith('ENV') && lines[i].split(' ')[1] == name) {
                version = lines[i].split(' ')[2]
                break
            }
        }
        script.echo "[DockerArtifactVersioning] Version from Docker environment variable ${name}: ${version}"
        return version.trim()
    }

    def getVersionFromDockerBaseImageTag(filePath) {
        def lines = script.readFile(filePath).split('\n')
        def version = null
        for (def i = 0; i < lines.size(); i++) {
            if (lines[i].startsWith('FROM') && lines[i].indexOf(':') > 0) {
                version = lines[i].split(':')[1]
                break
            }
        }
        script.echo "[DockerArtifactVersioning] Version from Docker base image tag: ${version}"
        return version.trim()
    }
}
