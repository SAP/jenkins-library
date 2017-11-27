package com.sap.piper

class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    private Map defaultValues

    private DefaultValueCache(Map defaultValues){
        this.defaultValues = defaultValues
    }

    static getInstance(){
        return instance
    }

    static createInstance(Map defaultValues){
        instance = new DefaultValueCache(defaultValues)
    }

    Map getDefaultValues(){
        return defaultValues
    }

    static reset(){
        instance = null
    }
}
