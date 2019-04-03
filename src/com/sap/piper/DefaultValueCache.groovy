package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

/*
 * Provides the default values in an immutable form. Maps, Sets, Lists
 * contained nested in the outermost Map are also immutable.
 *
 * java.util.Date instances are refused since they are not immutable.
 *
 * All other Objects we expect here (... basically from parsing json) are expected to be
 * immutable:
 *   - java.lang.String (in the unlikely case we see other CharSequences we convert to String)
 *   - java.lang.BigDecimal
 *   - java.lang.Integer
 *   - java.lang.Boolean
 */
@API
class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    private Map defaultValues

    private DefaultValueCache(Map defaultValues){
        this.defaultValues = defaultValues
    }

    @NonCPS
    static getInstance(){
        return instance
    }

    static createInstance(Map defaultValues){
        instance = new DefaultValueCache(
            immutable(defaultValues)
        )
    }

    @NonCPS
    Map getDefaultValues(){
        return defaultValues
    }

    static reset(){
        instance = null
    }


    private static Set immutable(Set s) {
        immutable(s, (Set)[])
    }

    private static List immutable(List l) {
        immutable(l, (List)[])
    }

    private static def immutable(Collection _in, Collection _out) {

        for(def e : _in) {
            if(e in List || e in Set || e in Map) {
                _out.add(immutable(e))
            } else if (e in CharSequence && ! (e in String)) {
                _out.add(e.toString())
            } else {
                typeCheck(e)
                _out.add(e)
            }
        }
        return _out.asImmutable()
    }

    private static Map immutable(Map m) {

        Map result = [:]

        for(def e : m.entrySet()) {

            if(e.value in Map) {
                result.put(e.key, immutable(e.value))
            } else if (e.value in List || e.value in Set) {
                result.put(e.key, immutable(e.value))
            } else if (e.value in CharSequence && ! (e.value in String)) {
                result.put(e.key, e.value.toString())
            } else {
                typeCheck(e.value)
                result.put(e.key, e.value)
            }
        }
        return result.asImmutable()
    }

        private static void typeCheck(def v) {
            if( ! (
                v in String ||
                v in BigDecimal ||
                v in Integer ||
                v in Boolean)

            ) throw new IllegalStateException("Unexpected type found: ${v}, type ${v.getClass().getName()}")

}

}
