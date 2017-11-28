package com.sap.piper

class ConfigurationHelper implements Serializable {

    private final Map config

    ConfigurationHelper(Map config = [:]){
        this.config = config
    }

    def getConfigProperty(key) {
        if (config[key] != null && config[key].class == String) {
            return config[key].trim()
        }
        return config[key]
    }

    def getConfigProperty(key, defaultValue) {
        def value = getConfigProperty(key)
        if (value == null) {
            return defaultValue
        }
        return value
    }

    def isPropertyDefined(key){

        def value = getConfigProperty(key)

        if(value == null){
            return false
        }

        if(value.class == String){
            return value?.isEmpty() == false
        }

        if(value){
            return true
        }

        return false
    }

    def getMandatoryProperty(key, defaultValue) {

        def paramValue = config[key]

        if (paramValue == null)
            paramValue = defaultValue

        if (paramValue == null)
            throw new Exception("ERROR - NO VALUE AVAILABLE FOR ${key}")
        return paramValue
    }

    def getMandatoryProperty(key) {
        return getMandatoryProperty(key, null)
    }
}
