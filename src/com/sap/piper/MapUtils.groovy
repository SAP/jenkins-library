package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class MapUtils implements Serializable {
    @NonCPS
    static boolean isMap(object){
        return object in Map
    }
    
    @NonCPS
    static Map merge(Map a, Map b, skipNull = true/*, override = true*/) {
        Map result = [:]
        
        a = a ?: [:]

        result.putAll(a)

        for(String key : b.keySet())
            //if(override || a[key] != null)
            if(isMap(b[key]))
                result[key] = merge(a[key], b[key])
            else if(b[key] != null || !skipNull)
                result[key] = b[key]
            // else: keep defaults value and omit null values from config
        return result
    }
}
