package com.sap.piper.versioning

class DockerArtifactVersioning extends ArtifactVersioning {
    protected DockerArtifactVersioning(script, configuration) {
        super(script, configuration)
    }

    def getVersion() {
        if(configuration.artifactType == 'appContainer' && configuration.dockerVersionSource == 'appVersion'){
            //replace + sign if available since + is not allowed in a Docker tag
            if (script.commonPipelineEnvironment.getArtifactVersion()){
                return script.commonPipelineEnvironment.getArtifactVersion().replace('+', '_')
            }else{
                throw new IllegalArgumentException("No artifact version available for 'dockerVersionSource: appVersion' -> executeBuild needs to run for the application artifact first to set the appVersion attribute.'")
            }
        } else if (configuration.dockerVersionSource == 'FROM') {
            def version = getVersionFromDockerBaseImageTag(configuration.filePath)
            if (version) {
                return  getVersionFromDockerBaseImageTag(configuration.filePath)
            } else {
                throw new IllegalArgumentException("No version information available in FROM statement")
            }
        } else {
            def version = getVersionFromDockerEnvVariable(configuration.filePath, configuration.dockerVersionSource)
            if (version) {
                return version
            } else {
                throw new IllegalArgumentException("ENV variable '${configuration.dockerVersionSource}' not found.")
            }
        }
    }

    @Override
    def setVersion(version) {
        def dockerVersionDir = (configuration.dockerVersionDir?dockerVersionDir:'')
        script.dir(dockerVersionDir) {
            script.writeFile file:'VERSION', text: version
        }
    }

    def getVersionFromDockerEnvVariable(filePath, envVarName) {
        def lines = script.readFile(filePath).split('\n')
        def version = ''
        for (def i = 0; i < lines.size(); i++) {
            if (lines[i].startsWith('ENV') && lines[i].split(' ')[1] == envVarName) {
                version = lines[i].split(' ')[2]
                break
            }
        }
        echo("Version from Docker environment variable ${envVarName}: ${version}")
        return version.trim()
    }

    def getVersionFromDockerBaseImageTag(filePath) {
        def lines = script.readFile(filePath).split('\n')
        def version = null
        for (def i = 0; i < lines.size(); i++) {
            if (lines[i].startsWith('FROM') && lines[i].indexOf(':') > 0) {
                def imageParts = lines[i].split(':')
                version = imageParts[imageParts.size()-1]
                if (version.contains('/')) {
                    script.error "[${getClass().getSimpleName()}] FROM statement does not contain an explicit image version: ${lines[i]} "
                }
                break
            }
        }
        echo("Version from Docker base image tag: ${version}")
        return version.trim()
    }
}
