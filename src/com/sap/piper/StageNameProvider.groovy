package com.sap.piper

@Singleton
class StageNameProvider implements Serializable {
    static final long serialVersionUID = 1L
    // static final String CENTRAL_BUILD = "Central Build";
    // static final String BUILD = "Build";

    /** Stores a feature toggle for defaulting to technical names in stages */
    boolean useTechnicalStageNames = false

    String getStageName(Script script, Map parameters, Script step) {
        String stageName = null
        if (parameters.stageName in CharSequence) {
            stageName = parameters.stageName
            stageName = replaceCentralBuild(stageName);
            return stageName
        }
        if (this.useTechnicalStageNames) {
            String technicalStageName = getTechnicalStageName(step)
            if (technicalStageName) {
                return technicalStageName
            }
        }
        if (stageName == null) {
            stageName = script.env.STAGE_NAME
            stageName = replaceCentralBuild(stageName);
        }
        return stageName
    }

    private String replaceCentralBuild(String stageName) {
        return CENTRAL_BUILD.equals(stageName) ? BUILD : stageName;
    }

    static String getTechnicalStageName(Script step) {
        try {
            return step.TECHNICAL_STAGE_NAME
        } catch (Throwable ignored) {
        }
        return null
    }
}
