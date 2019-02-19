package com.sap.piper

import hudson.AbortException

class ContainerUtils implements Serializable {

    private static Script script

    ContainerUtils(Script script) {
        this.script = script
    }

    public boolean withDockerDeamon() {
        def returnCode = script.sh script: 'docker ps -q > /dev/null', returnStatus: true
        return (returnCode == 0)
    }

    public boolean onKubernetes() {
        return (Boolean.valueOf(script.env.ON_K8S) || (script.env.jaas_owner != null))
    }

    public void saveImage(filePath, dockerImage, dockerRegistryUrl = '') {
        def dockerRegistry = dockerRegistryUrl ? "${getRegistryFromUrl(dockerRegistryUrl)}/" : ''
        def imageFullName = dockerRegistry + dockerImage
        if (withDockerDeamon()) {
            if (dockerRegistry) {
                script.docker.withRegistry(dockerRegistryUrl) {
                    script.sh "docker pull ${imageFullName} && docker save --output ${filePath} ${imageFullName}"
                }
            } else {
                script.sh "docker pull ${imageFullName} && docker save --output ${filePath} ${imageFullName}"
            }
        } else {
            try {
                //assume that we are on Kubernetes
                //needs to run inside an existing pod in order to not move around heavy images
                skopeoSaveImage(imageFullName, dockerImage, filePath)
            } catch (err) {
                throw new AbortException('No Kubernetes container provided for running Skopeo ...')
            }
        }
    }

    private void skopeoSaveImage(imageFullName, dockerImage, filePath) {
        script.sh "skopeo copy --src-tls-verify=false docker://${imageFullName} docker-archive:${filePath}:${dockerImage}"
    }

    private void skopeoMoveImage(sourceImageFullName, targetImageFullName, targetUserId, targetPassword) {
        script.sh "skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=${BashUtils.quoteAndEscape(targetUserId)}:${BashUtils.quoteAndEscape(targetPassword)} docker://${sourceImageFullName} docker://${targetImageFullName}"
    }

    public String getRegistryFromUrl(dockerRegistryUrl) {
        return dockerRegistryUrl.split(/^https?:\/\//)[1]
    }

    public String getProtocolFromUrl(dockerRegistryUrl) {
        return dockerRegistryUrl.split(/:\/\//)[0]
    }

    public String getNameFromImageUrl(imageUrl) {

        def imageNameAndTag

        //remove digest if present
        imageUrl = imageUrl.split('@')[0]

        //remove registry part if present
        def pattern = /\.(?:[^\/]*)\/(.*)/
        def matcher = imageUrl =~ pattern
        if (matcher.size() == 0) {
            imageNameAndTag = imageUrl
        } else {
            imageNameAndTag = matcher[0][1]
        }

        //remove tag if present
        return imageNameAndTag.split(':')[0]
    }
}

