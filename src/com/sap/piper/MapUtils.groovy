package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class MapUtils implements Serializable {
    @NonCPS
    static isMap(object){
        return object in Map
    }

    @NonCPS
    static fromList(List list){
        Map map = [:]
        for(String key : list){
            map.put(key, null)
        }
        return map
    }
}
