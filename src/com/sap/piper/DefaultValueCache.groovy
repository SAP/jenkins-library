package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

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
        instance = new DefaultValueCache(defaultValues)
    }

    @NonCPS
    Map getDefaultValues(){
        return defaultValues
    }

    static reset(){
        instance = null
    }
}
