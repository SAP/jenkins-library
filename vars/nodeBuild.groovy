
void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
        sh '''npm run build'''
    }
}
