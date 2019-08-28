package com.sap.piper.k8s

class SidecarUtils {


    static waitForSidecarReadyOnDocker(String containerId, String command, Script script){
        String dockerCommand = "docker exec ${containerId} ${command}"
        waitForSidecarReady(dockerCommand, script)
    }

    static waitForSidecarReadyOnKubernetes(String containerName, String command, Script script){
        script.container(name: containerName){
            waitForSidecarReady(command, script)
        }
    }

    static waitForSidecarReady(String command, Script script){
        int sleepTimeInSeconds = 10
        int timeoutInSeconds = 5 * 60
        int maxRetries = timeoutInSeconds / sleepTimeInSeconds
        int retries = 0
        while(true){
            script.echo "Waiting for sidecar container"
            String status = script.sh script:command, returnStatus:true
            if(status == "0") return
            if(retries > maxRetries){
                script.error("Timeout while waiting for sidecar container to be ready")
            }

            sleep sleepTimeInSeconds
            retries++
        }
    }
}
