package com.sap.piper

class SidecarUtils implements Serializable {

    private static Script script

    SidecarUtils(Script script) {
        this.script = script
    }

    void waitForSidecarReadyOnDocker(String containerId, String command) {
        String dockerCommand = "docker exec ${containerId} ${command}"
        waitForSidecarReady(dockerCommand)
    }

    void waitForSidecarReadyOnKubernetes(String containerName, String command) {
        script.container(name: containerName) {
            waitForSidecarReady(command)
        }
    }

    void waitForSidecarReady(String command) {
        int sleepTimeInSeconds = 10
        int timeoutInSeconds = 5 * 60
        int maxRetries = timeoutInSeconds / sleepTimeInSeconds
        int retries = 0
        while (true) {
            script.echo "Waiting for sidecar container"
            String status = script.sh script: command, returnStatus: true
            if (status == "0") {
                return
            }
            if (retries > maxRetries) {
                script.error("Timeout while waiting for sidecar container to be ready")
            }

            sleep sleepTimeInSeconds
            retries++
        }
    }
}
