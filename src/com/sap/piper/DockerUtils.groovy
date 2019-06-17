package com.sap.piper

class DockerUtils implements Serializable {

    private static Script script

    DockerUtils(Script script) {
        this.script = script
    }

    public boolean withDockerDeamon() {
        def returnCode = script.sh script: 'docker ps -q > /dev/null', returnStatus: true
        return (returnCode == 0)
    }

    public boolean onKubernetes() {
        return (Boolean.valueOf(script.env.ON_K8S))
    }

    public String getRegistryFromUrl(dockerRegistryUrl) {
        return dockerRegistryUrl.split(/^https?:\/\//)[1]
    }

    public String getProtocolFromUrl(dockerRegistryUrl) {
        return dockerRegistryUrl.split(/:\/\//)[0]
    }

    public void moveImage(Map source, Map target) {
        //expects source/target in the format [image: '', registryUrl: '', credentialsId: '']
        def sourceDockerRegistry = source.registryUrl ? "${getRegistryFromUrl(source.registryUrl)}/" : ''
        def sourceImageFullName = sourceDockerRegistry + source.image
        def targetDockerRegistry = target.registryUrl ? "${getRegistryFromUrl(target.registryUrl)}/" : ''
        def targetImageFullName = targetDockerRegistry + target.image

        if (!withDockerDeamon()) {
            script.withCredentials([script.usernamePassword(
                credentialsId: target.credentialsId,
                passwordVariable: 'password',
                usernameVariable: 'userid'
            )]) {
                skopeoMoveImage(sourceImageFullName, targetImageFullName, script.userid, script.password)
            }
        }
        //else not yet implemented here - available directly via pushToDockerRegistry

    }

    private void skopeoMoveImage(sourceImageFullName, targetImageFullName, targetUserId, targetPassword) {
        script.sh "skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=${BashUtils.quoteAndEscape(targetUserId)}:${BashUtils.quoteAndEscape(targetPassword)} docker://${sourceImageFullName} docker://${targetImageFullName}"
    }


}
