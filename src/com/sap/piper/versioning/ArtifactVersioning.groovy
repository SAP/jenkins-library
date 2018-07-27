package com.sap.piper.versioning

abstract class ArtifactVersioning implements Serializable {

    final protected script
    final protected Map configuration

    protected ArtifactVersioning(script, configuration) {
        this.script = script
        this.configuration = configuration
    }

    public static getArtifactVersioning(buildTool, script, configuration) {
        switch (buildTool) {
            case 'mta':
                return new MtaArtifactVersioning(script, configuration)
            case 'maven':
                return new MavenArtifactVersioning(script, configuration)
            case 'docker':
                return new DockerArtifactVersioning(script, configuration)
            default:
                throw new IllegalArgumentException("No versioning implementation for buildTool: ${buildTool} available.")
        }
    }

    abstract setVersion(version)
    abstract getVersion()

    protected echo(msg){
        script.echo("[${this.getClass().getSimpleName()}] ${msg}")
    }
}
