package com.sap.piper.k8s

import com.sap.piper.API

@API
@Singleton
class ContainerMap {
    private Map containerMap = null

    Map getMap() {
        if (containerMap == null) {
            containerMap = [:]
        }
        return containerMap
    }

    void setMap(Map containersMap) {
        containerMap = containersMap
    }
}
