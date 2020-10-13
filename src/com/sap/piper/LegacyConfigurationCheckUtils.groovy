package com.sap.piper

class LegacyConfigurationCheckUtils{
    static void checkConfiguration(Script script, Map configChanges) {
        if (configChanges?.removedOrReplacedConfigKeys) {
            checkForRemovedOrReplacedConfigKeys(script, configChanges.removedOrReplacedConfigKeys)
        }

        if (configChanges?.removedOrReplacedSteps) {
            checkForRemovedOrReplacedSteps(script, configChanges.removedOrReplacedSteps)
        }

        if (configChanges?.removedOrReplacedStages) {
            checkForRemovedOrReplacedStages(script, configChanges.removedOrReplacedStages)
        }

        if (configChanges?.parameterTypeChanged) {
            checkForParameterTypeChanged(script, configChanges.parameterTypeChanged)
        }

        if (configChanges?.renamedNpmScript) {
            checkForRenamedNpmScripts(script, configChanges.renamedNpmScript)
        }
    }

    static void checkForRemovedOrReplacedConfigKeys(Script script, Map configChanges) {
        configChanges.each {oldConfigKey, changes ->
            List steps = changes?.steps ?: []
            List stages = changes?.stages ?: []
            Boolean general = changes?.general ?: false
            Boolean postAction = changes?.postAction ?: false

            Boolean warnInsteadOfError = changes?.warnInsteadOfError ?: false
            String customMessage = changes?.customMessage ?: ""
            String newConfigKey = changes?.newConfigKey ?: ""

            if (newConfigKey) {
                customMessage = "Please use the parameter ${newConfigKey} instead. " + customMessage
            }

            for (int i = 0; i < steps.size(); i++) {
                Map config = loadEffectiveStepConfig(script, steps[i])
                if (config.containsKey(oldConfigKey)) {
                    String errorMessage = "Your pipeline configuration contains the configuration key ${oldConfigKey} for the step ${steps[i]}. " +
                        "This configuration option was removed. " + customMessage
                    if (warnInsteadOfError) {
                        addPipelineWarning(script, "Deprecated configuration key ${oldConfigKey}", errorMessage)
                    } else {
                        script.error(errorMessage)
                    }
                }
            }

            for (int i = 0; i < stages.size(); i++) {
                Map config = loadEffectiveStageConfig(script, stages[i])
                if (config.containsKey(oldConfigKey)) {
                    String errorMessage = "Your pipeline configuration contains the configuration key ${oldConfigKey} for the stage ${stages[i]}. " +
                        "This configuration option was removed. " + customMessage
                    if (warnInsteadOfError) {
                        addPipelineWarning(script, "Deprecated configuration key ${oldConfigKey}", errorMessage)
                    } else {
                        script.error(errorMessage)
                    }
                }
            }

            if (general) {
                Map config = loadEffectiveGeneralConfig(script)
                if (config.containsKey(oldConfigKey)) {
                    String errorMessage = "Your pipeline configuration contains the configuration key ${oldConfigKey} in the general section. " +
                        "This configuration option was removed. " + customMessage
                    if (warnInsteadOfError) {
                        addPipelineWarning(script, "Deprecated configuration key ${oldConfigKey}", errorMessage)
                    } else {
                        script.error(errorMessage)
                    }
                }
            }

            if (postAction) {
                Map config = loadEffectivePostActionConfig(script)
                if (config.containsKey(oldConfigKey)) {
                    String errorMessage = "Your pipeline configuration contains the configuration key ${oldConfigKey} in the postActions section. " +
                        "This configuration option was removed. " + customMessage
                    if (warnInsteadOfError) {
                        addPipelineWarning(script, "Deprecated configuration key ${oldConfigKey}", errorMessage)
                    } else {
                        script.error(errorMessage)
                    }
                }
            }
        }
    }

    static void checkForRemovedOrReplacedSteps(Script script, Map configChanges) {
        configChanges.each {oldConfigKey, changes ->
            Boolean onlyCheckProjectConfig = changes?.onlyCheckProjectConfig ?: false
            String customMessage = changes?.customMessage ?: ""
            String newStepName = changes?.newStepName ?: ""

            if (newStepName) {
                customMessage = "Please configure the step ${newStepName} instead. " + customMessage
            }

            Map config
            if (onlyCheckProjectConfig) {
                config = ConfigurationLoader.stepConfiguration(script, oldConfigKey)
            } else {
                config = loadEffectiveStepConfig(script, oldConfigKey)
            }

            if (config) {
                script.error("Your pipeline configuration contains configuration for the step ${oldConfigKey}. " +
                    "This step has been removed. " + customMessage)
            }
        }
    }

    static void checkForRemovedOrReplacedStages(Script script, Map configChanges) {
        configChanges.each {oldConfigKey, changes ->
            String customMessage = changes?.customMessage ?: ""
            String newStageName = changes?.newStageName ?: ""

            if (newStageName) {
                customMessage = "Please configure the stage ${newStageName} instead. " + customMessage
            }

            if (loadEffectiveStageConfig(script, oldConfigKey)) {
                script.error("Your pipeline configuration contains configuration for the stage ${oldConfigKey}. " +
                    "This stage has been removed. " + customMessage)
            }
        }
    }

    static void checkForParameterTypeChanged(Script script, Map configChanges) {
        configChanges.each { oldConfigKey, changes ->
            String oldType = changes?.oldType ?: ""
            String newType = changes?.newType ?: ""
            List steps = changes?.steps ?: []
            List stages = changes?.stages ?: []
            Boolean general = changes?.general ?: false
            String customMessage = changes?.customMessage ?: ""

            if (oldType != "String") {
                script.echo ("Your legacy config settings contain an entry for parameterTypeChanged with the key ${oldConfigKey} with the unsupported type ${oldType}. " +
                    "Currently only the type 'String' is supported.")
                return
            }

            for (int i = 0; i < steps.size(); i++) {
                Map config = loadEffectiveStepConfig(script, steps[i])
                if (config.containsKey(oldConfigKey)) {
                    if (oldType == "String" && config.get(oldConfigKey) instanceof String) {
                        script.error("Your pipeline configuration contains the configuration key ${oldConfigKey} for the step ${steps[i]}. " +
                            "The type of this configuration parameter was changed from ${oldType} to ${newType}. " + customMessage)
                    }
                }
            }

            for (int i = 0; i < stages.size(); i++) {
                Map config = loadEffectiveStageConfig(script, stages[i])
                if (config.containsKey(oldConfigKey)) {
                    if (oldType == "String" && config.get(oldConfigKey) instanceof String) {
                        script.error("Your pipeline configuration contains the configuration key ${oldConfigKey} for the stage ${stages[i]}. " +
                            "The type of this configuration parameter was changed from ${oldType} to ${newType}. " + customMessage)
                    }
                }
            }

            if (general) {
                Map config = loadEffectiveGeneralConfig(script)
                if (config.containsKey(oldConfigKey)) {
                    if (oldType == "String" && config.get(oldConfigKey) instanceof String) {
                        script.error("Your pipeline configuration contains the configuration key ${oldConfigKey} in the general section. " +
                            "The type of this configuration parameter was changed from ${oldType} to ${newType}. " + customMessage)
                    }
                }
            }
        }
    }

    static void checkForRenamedNpmScripts(Script script, Map configChanges) {
        configChanges.each { oldScriptName, changes ->
            String newScriptName = changes?.newScriptName ?: ""
            Boolean warnInsteadOfError = changes?.warnInsteadOfError ?: false
            String customMessage = changes?.customMessage ?: ""

            String packageJsonWithScript = findPackageWithScript(script, oldScriptName)
            if (packageJsonWithScript) {
                String errorMessage = "Your package.json file ${packageJsonWithScript} contains an npm script using the deprecated name ${oldScriptName}. " +
                    "Please rename the script to ${newScriptName}, since the script ${oldScriptName} will not be executed by the pipeline anymore. " + customMessage
                if (warnInsteadOfError) {
                    addPipelineWarning(script, "Deprecated npm script ${oldScriptName}", errorMessage)
                } else {
                    script.error(errorMessage)
                }
            }
        }
    }

    private static String findPackageWithScript(Script script, String scriptName) {
        List packages = script.findFiles(glob: '**/package.json', excludes: '**/node_modules/**')

        for (int i = 0; i < packages.size(); i++) {
            String packageJsonPath = packages[i].path
            Map packageJson = script.readJSON file: packageJsonPath
            Map npmScripts = packageJson?.scripts ?: [:]
            if (npmScripts.get(scriptName)) {
                return packageJsonPath
            }
        }
        return ""
    }

    private static Map loadEffectiveStepConfig(Script script, String stepName) {
        return MapUtils.merge(ConfigurationLoader.defaultStepConfiguration(script, stepName), ConfigurationLoader.stepConfiguration(script, stepName))
    }

    private static Map loadEffectiveStageConfig(Script script, String stageName) {
        return MapUtils.merge(ConfigurationLoader.defaultStageConfiguration(script, stageName), ConfigurationLoader.stageConfiguration(script, stageName))
    }

    private static Map loadEffectiveGeneralConfig(Script script) {
        return MapUtils.merge(ConfigurationLoader.defaultGeneralConfiguration(script), ConfigurationLoader.generalConfiguration(script))
    }

    private static Map loadEffectivePostActionConfig(Script script) {
        Map defaultPostActionConfig = DefaultValueCache.getInstance()?.getDefaultValues()?.get("postActions") ?: [:]
        Map projectPostActionConfig = script?.commonPipelineEnvironment?.configuration?.postActions ?: [:]
        return MapUtils.merge(defaultPostActionConfig, projectPostActionConfig)
    }

    static void addPipelineWarning(Script script, String heading, String message) {
        script.echo '[WARNING] ' + message
        script.addBadge(icon: "warning.gif", text: message)

        String html =
            """
            <h2>$heading</h2>
            <p>$message</p>
            """

        script.createSummary(icon: "warning.gif", text: html)
    }
}
