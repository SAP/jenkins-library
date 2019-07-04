package com.sap.piper.k8s

import com.cloudbees.groovy.cps.NonCPS

class SystemEnv implements Serializable {
    static final long serialVersionUID = 1L

    private Map env = new HashMap<String, String>()

    Set<String> envNames = [
        'HTTP_PROXY',
        'HTTPS_PROXY',
        'NO_PROXY',
        'http_proxy',
        'https_proxy',
        'no_proxy',
        'ON_K8S'
    ]

    SystemEnv() {
        fillMap()
    }

    String get(String key) {
        return env.get(key)
    }

    Map getEnv() {
        return env
    }

    String remove(String key) {
        return env.remove(key)
    }

    @NonCPS
    private void fillMap() {
        for (String name in envNames) {
            if(System.getenv(name)){
                env.put(name,System.getenv(name))
            }
        }
    }
}
