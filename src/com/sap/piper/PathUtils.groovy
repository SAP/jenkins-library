package com.sap.piper

import groovy.text.SimpleTemplateEngine

class PathUtils implements Serializable {
    static final long serialVersionUID = 1L

    static String fillPathTemplate(Script script, String templateString) {
        if(!templateString) {
            return templateString
        }
        Map templateValues = [
            workspaceRoot: script.env.WORKSPACE
        ]
        return new SimpleTemplateEngine().createTemplate(templateString).make(templateValues).toString()
    }

    static Map replacePathInConfiguration(Script script, Map configuration, Set keysContainingAPath){
        configuration = MapUtils.deepCopy(configuration)
        for(int i=0; i<keysContainingAPath.size(); i++){
            String key = keysContainingAPath.getAt(i)
            configuration[key] = fillPathTemplate(script, configuration[key])
        }
        return configuration
    }
}
