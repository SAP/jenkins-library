package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    enum Level { DEFAULTS }

    private Map<Level, Map> configurations = [:]

    private DefaultValueCache(Map defaultValues){
        this.configurations.put(Level.DEFAULTS, defaultValues)
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
        return configurations.get(Level.DEFAULTS)
    }

    static reset(){
        instance = null
    }
}
