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
            else {
                // do not update the map while it is traversed. Depending
                // on the map implementation the behavior is undefined.
                updates.put(e.getKey(), strategy(e.getValue()))
            }
        }
        m.putAll(updates)
    }

    @NonCPS
    static private def getByPath(Map m, def key) {
        List path = key in CharSequence ? key.tokenize('/') : key

        def value = m.get(path.head())

        if (path.size() == 1) return value
        if (value in Map) return getByPath(value, path.tail())

        return null
    }

    /*
     * Provides a new map with the same content like the original map.
     * Nested Collections and Maps are copied. Values with are not
     * Collections/Maps are not copied/cloned.
     * &lt;paranoia&gt;&/ltThe keys are also not copied/cloned, even if they are
     * Maps or Collections;paranoia&gt;
     */
    @NonCPS
    static deepCopy(Map original) {
        Map copy = [:]
        for (def e : original.entrySet()) {
            if(e.value == null) {
                copy.put(e.key, e.value)
            } else {
                copy.put(e.key, deepCopy(e.value))
            }
        }
        copy
    }

    @NonCPS
    /* private */ static deepCopy(Set original) {
        Set copy = []
        for(def e : original)
            copy << deepCopy(e)
        copy
    }

    @NonCPS
    /* private */ static deepCopy(List original) {
        List copy = []
        for(def e : original)
            copy << deepCopy(e)
        copy
    }

    /*
     * In fact not a copy, but a catch all for everything not matching
     * with the other signatures
     */
    @NonCPS
    /* private */ static deepCopy(def original) {
        original
    }
}
