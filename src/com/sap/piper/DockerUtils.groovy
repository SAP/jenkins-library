package com.sap.piper

class DockerUtils implements Serializable {

    private static Script script

    DockerUtils(Script script) {
        this.script = script
    }

    public boolean withDockerDaemon() {
        def returnCode = script.sh script: 'docker ps -q > /dev/null', returnStatus: true
        return (returnCode == 0)
    }

    public boolean onKubernetes() {
        return (Boolean.valueOf(script.env.ON_K8S))
    }

    public String getRegistryFromUrl(dockerRegistryUrl) {
        URL url = new URL(dockerRegistryUrl)
        return "${url.getHost()}${(url.getPort() != -1) ? ':' + url.getPort() : ''}"
    }

    public String getProtocolFromUrl(dockerRegistryUrl) {
        URL url = new URL(dockerRegistryUrl)
        return url.getProtocol()

        //return dockerRegistryUrl.split(/:\/\//)[0]
    }

    public void moveImage(Map source, Map target) {
        //expects source/target in the format [image: '', registryUrl: '', credentialsId: '']
        def sourceDockerRegistry = source.registryUrl ? "${getRegistryFromUrl(source.registryUrl)}/" : ''
        def sourceImageFullName = sourceDockerRegistry + source.image
        def targetDockerRegistry = target.registryUrl ? "${getRegistryFromUrl(target.registryUrl)}/" : ''
        def targetImageFullName = targetDockerRegistry + target.image

        if (!withDockerDaemon()) {
            if (source.credentialsId) {
                script.withCredentials([
                    script.usernamePassword(credentialsId: source.credentialsId, passwordVariable: 'src_password', usernameVariable: 'src_userid'), 
                    script.usernamePassword(credentialsId: target.credentialsId, passwordVariable: 'password', usernameVariable: 'userid')
                ]) {
                    skopeoMoveImage(sourceImageFullName, script.src_userid, script.src_password, targetImageFullName, script.userid, script.password)
                }
            } else {
                script.withCredentials([
                    script.usernamePassword(credentialsId: target.credentialsId, passwordVariable: 'password', usernameVariable: 'userid')
                ]) {
                    skopeoMoveImage(sourceImageFullName, '', '', targetImageFullName, script.userid, script.password)
                }
            }
        }
        //else not yet implemented here - available directly via containerPushToRegistry

    }

    private void skopeoMoveImage(sourceImageFullName, sourceUserId, sourcePassword, targetImageFullName, targetUserId, targetPassword) {
        if (sourceUserId && sourcePassword) {
            script.sh "skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=${BashUtils.quoteAndEscape(sourceUserId)}:${BashUtils.quoteAndEscape(sourcePassword)} --dest-tls-verify=false --dest-creds=${BashUtils.quoteAndEscape(targetUserId)}:${BashUtils.quoteAndEscape(targetPassword)} docker://${sourceImageFullName} docker://${targetImageFullName}"
        } else {
            script.sh "skopeo copy --multi-arch=all --src-tls-verify=false --dest-tls-verify=false --dest-creds=${BashUtils.quoteAndEscape(targetUserId)}:${BashUtils.quoteAndEscape(targetPassword)} docker://${sourceImageFullName} docker://${targetImageFullName}"
        }
    }
}
