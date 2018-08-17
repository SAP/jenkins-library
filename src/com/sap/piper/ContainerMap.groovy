package com.sap.piper

class ContainerMap {
    private static final ContainerMap instance = new ContainerMap();
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
