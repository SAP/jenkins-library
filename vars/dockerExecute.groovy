import com.cloudbees.groovy.cps.NonCPS

def call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: 'dockerExecute', stepParameters: parameters){
        def dockerImage = parameters.get('dockerImage', '')
        Map dockerEnvVars = parameters.get('dockerEnvVars', [:])
        def dockerOptions = parameters.get('dockerOptions', '')
        Map dockerVolumeBind = parameters.get('dockerVolumeBind', [:])

        if(dockerImage?.isEmpty()){
            echo '[dockerExecute] No Docker image provided - running on local environment.'
            body()
        }else{
            def image = docker.image(dockerImage)
            image.pull()
            image.inside(getDockerOptions(dockerEnvVars, dockerVolumeBind, dockerOptions)) {
                body()
            }
        }
    }
}

/**
 * Returns a string with docker options containing
 * environment variables (if set).
 * Possible to extend with further options.
 * @param dockerEnvVars Map with environment variables
 */
@NonCPS
private getDockerOptions(Map dockerEnvVars, Map dockerVolumeBind, def dockerOptions) {
    def specialEnvironments = [
        'http_proxy',
        'https_proxy',
        'no_proxy',
        'HTTP_PROXY',
        'HTTPS_PROXY',
        'NO_PROXY'
    ]
    def options = ""
    if (dockerEnvVars) {
        for (String k : dockerEnvVars.keySet()) {
            options += " --env ${k}=" + dockerEnvVars[k].toString()
        }
    }

    for (String envVar : specialEnvironments) {
        if (dockerEnvVars == null || !dockerEnvVars.containsKey('envVar')) {
            options += " --env ${envVar}"
        }
    }

    if (dockerVolumeBind) {
        for (String k : dockerVolumeBind.keySet()) {
            options += " --volume ${k}:" + dockerVolumeBind[k].toString()
        }
    }

    if (dockerOptions) {
        options += " ${dockerOptions}"
    }
    return options
}
