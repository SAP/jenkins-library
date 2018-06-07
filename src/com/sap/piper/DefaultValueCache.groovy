package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    // Contains defaults values provided by this library itself
    private Map defaultValues

    // intended for describing e.g. the system landscape on customer side
    private Map customDefaultValues

    private DefaultValueCache(Map customeDefaultValues, Map customDefaultValues){
        this.defaultValues = defaultValues
        this.customDefaultValues = customDefaultValues ?: [:]
    }

    @NonCPS
    static getInstance(){
        return instance
    }

    static createInstance(Map defaultValues, Map customDefaultValues){
        instance = new DefaultValueCache(defaultValues, customDefaultValues)
    }

    @NonCPS
    Map getDefaultValues(){
        return defaultValues
    }

    @NonCPS
    Map getCustomDefaultValues(){
        return customDefaultValues
    }

    static reset(){
        instance = null
    }
}
