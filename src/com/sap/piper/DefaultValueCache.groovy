package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    enum Level { DEFAULTS, PROJECT }

    private Map<Level, Map> configurations = [:]

    private DefaultValueCache(Map defaultValues, Map projectConfiguration){
        this.configurations.put(Level.DEFAULTS, defaultValues)
        this.configurations.put(Level.PROJECT, projectConfiguration)
    }

    @NonCPS
    static getInstance(){
        return instance
    }

    static createInstance(Map defaultValues, Map projectConfiguration){
        instance = new DefaultValueCache(defaultValues, projectConfiguration)
    }

    @NonCPS
    Map getDefaultValues(){
        return configurations.get(Level.DEFAULTS)
    }

    @NonCPS
    Map getProjectConfiguration(){
        return configurations.get(Level.PROJECT)
    }


    static reset(){
        instance = null
    }
}
