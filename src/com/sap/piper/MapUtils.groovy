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

    /**
     * @param m The map to which the changed denoted by closure <code>strategy</code>
     *        should be applied.
     *        The strategy is also applied to all sub-maps contained as values
     *        in <code>m</code> in a recursive manner.
     * @param strategy Strategy applied to all non-map entries
     */
    @NonCPS
    static void traverse(Map m, Closure strategy) {

        def updates = [:]
        for(def e : m.entrySet()) {
            if(isMap(e.value)) {
                traverse(e.getValue(), strategy)
            }
            else
                // do not update the map while it is traversed. Depending
                // on the map implementation the behavior is undefined.
                updates.put(e.key, strategy(e.value))
        }
        m.putAll(updates)
    }
}
