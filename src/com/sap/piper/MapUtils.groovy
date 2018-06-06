package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class MapUtils implements Serializable {
    @NonCPS
    static boolean isMap(object){
        return object in Map
    }

    @NonCPS
    static Map pruneNulls(Map m) {

        Map result = [:]

        m = m ?: [:]

        for(def e : m.entrySet())
            if(isMap(e.value))
                result[e.key] = pruneNulls(e.value)
            else if(e.value != null)
                result[e.key] = e.value
        return result
    }


    @NonCPS
    static Map merge(Map base, Map overlay) {

        Map result = [:]

        base = base ?: [:]

        result.putAll(base)

        for(def e : overlay.entrySet())
            result[e.key] = isMap(e.value) ? merge(base[e.key], e.value) : e.value

        return result
    }
}
