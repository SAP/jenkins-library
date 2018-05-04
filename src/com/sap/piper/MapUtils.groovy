package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class MapUtils implements Serializable {
    @NonCPS
    static boolean isMap(object){
        return object in Map
    }
}
