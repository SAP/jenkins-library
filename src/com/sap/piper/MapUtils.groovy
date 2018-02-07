package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class MapUtils implements Serializable {
    @NonCPS
    static isMap(object){
        return object in Map
    }
}
