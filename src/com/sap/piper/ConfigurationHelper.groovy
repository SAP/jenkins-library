package com.sap.piper

class ConfigurationHelper implements Serializable {

    private final Map config

    ConfigurationHelper(Map config = [:]){
        this.config = config
    }

    def getConfigProperty(property) {
        if (config[property] != null && config[property].class == String) {
            return config[property].trim()
        }
        return config[property]
    }

    def getConfigProperty(property, defaultValue) {
        def value = getConfigProperty(property)
        if (value == null) {
            return defaultValue
        }
        return value
    }

    def isPropertyDefined(property){

        def value = getConfigProperty(property)

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
