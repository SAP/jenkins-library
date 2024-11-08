package com.sap.piper

@Singleton
class StageNameProvider implements Serializable {
    static final long serialVersionUID = 1L

    /** Stores a feature toggle for defaulting to technical names in stages */
    boolean useTechnicalStageNames = false

    String getStageName(Script script, Map parameters, Script step) {
        if (parameters.stageName in CharSequence) {
            if (parameters.stageName == 'Central Build'){
                return 'Build'
            }
            return stageName
        }
        if (this.useTechnicalStageNames) {
            String technicalStageName = getTechnicalStageName(step)
            if (technicalStageName) {
                return technicalStageName
            }
        } 
        if (script.env.STAGE_NAME == 'Central Build'){
            return = 'Build'
        }
        return script.env.STAGE_NAME
    }

    static String getTechnicalStageName(Script step) {
        try {
            return step.TECHNICAL_STAGE_NAME
        } catch (Throwable ignored) {
        }
        return null
    }
}
