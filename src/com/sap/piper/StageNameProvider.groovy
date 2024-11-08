package com.sap.piper

@Singleton
class StageNameProvider implements Serializable {
    static final long serialVersionUID = 1L

    /** Stores a feature toggle for defaulting to technical names in stages */
    boolean useTechnicalStageNames = false

    String getStageName(Script script, Map parameters, Script step) {
        if (parameters.stageName in CharSequence) {
            stageName = parameters.stageName
            if (stageName == 'Central Build'){
                stageName = 'Build'
            }
            return stageName
        }
        if (this.useTechnicalStageNames) {
            String technicalStageName = getTechnicalStageName(step)
            if (technicalStageName) {
                return technicalStageName
            }
        }
        stageName = script.env.STAGE_NAME
        if (stageName == 'Central Build'){
            stageName = 'Build'
        }
        return stageName
    }

    static String getTechnicalStageName(Script step) {
        try {
            return step.TECHNICAL_STAGE_NAME
        } catch (Throwable ignored) {
        }
        return null
    }
}
