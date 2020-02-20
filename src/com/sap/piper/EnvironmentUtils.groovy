package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import java.nio.file.Files
import java.nio.file.Paths

class EnvironmentUtils implements Serializable {
    static boolean cxServerDirectoryExists() {
        return Files.isDirectory(Paths.get('/var/cx-server/'));
    }

    @NonCPS
    static String getDockerFile(String serverCfgAsString) {
        String result = 'not_found'
        serverCfgAsString.splitEachLine("=") { items ->
            if (items[0].trim() == 'docker_image') {
                result = items[1].trim().replaceAll('"', '')
            }
        }
        return result
    }
}
