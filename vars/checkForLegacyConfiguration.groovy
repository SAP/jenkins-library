import com.sap.piper.ConfigurationLoader
import com.sap.piper.DefaultValueCache
import com.sap.piper.MapUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    List errors = []
    def script = checkScript(this, parameters) ?: this
    def configChanges = parameters.legacyConfigSettings

    if (configChanges?.removedOrReplacedConfigKeys) {
        errors.addAll(checkForRemovedOrReplacedConfigKeys(script, configChanges.removedOrReplacedConfigKeys))
    }

    if (configChanges?.requiredConfigKeys) {
        errors.addAll(checkForMissingConfigKeys(script, configChanges.requiredConfigKeys))
    }

    if (configChanges?.removedOrReplacedSteps) {
        errors.addAll(checkForRemovedOrReplacedSteps(script, configChanges.removedOrReplacedSteps))
    }

    if (configChanges?.removedOrReplacedStages) {
        errors.addAll(checkForRemovedOrReplacedStages(script, configChanges.removedOrReplacedStages))
    }

    if (configChanges?.parameterTypeChanged) {
        errors.addAll(checkForParameterTypeChanged(script, configChanges.parameterTypeChanged))
    }

    if (configChanges?.renamedNpmScript) {
        errors.addAll(checkForRenamedNpmScripts(script, configChanges.renamedNpmScript))
    }

    if (errors) {
        String message = ""
        if (errors.size() > 1) {
            message += "Your pipeline configuration file contains the following errors:\n"
        }
        for (error in errors) {
            message += error
            message += "\n"
        }
        message += "Failing pipeline due to configuration errors. Please see log output above."
        script.error(message)
    }
}

static List checkForRemovedOrReplacedConfigKeys(Script script, Map configChanges) {
    List errors = []
    configChanges.each { oldConfigKey, changes ->
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
                    errors.add(errorMessage)
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
                    errors.add(errorMessage)
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
                    errors.add(errorMessage)
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
                    errors.add(errorMessage)
                }
            }
        }
    }
    return errors
}

static List checkForMissingConfigKeys(Script script, Map configChanges) {
    List errors = []
    configChanges.each { requiredConfigKey, changes ->
        List steps = changes?.steps ?: []
        List stages = changes?.stages ?: []
        Boolean general = changes?.general ?: false
        Boolean postAction = changes?.postAction ?: false

        Boolean warnInsteadOfError = changes?.warnInsteadOfError ?: false
        String customMessage = changes?.customMessage ?: ""

        for (int i = 0; i < steps.size(); i++) {
            Map config = loadEffectiveStepConfig(script, steps[i])
            if (!config.containsKey(requiredConfigKey)) {
                String errorMessage = "Your pipeline configuration does not contain the configuration " +
                    "key ${requiredConfigKey} for the step ${steps[i]}. " + customMessage
                if (warnInsteadOfError) {
                    addPipelineWarning(script, "Missing configuration key ${requiredConfigKey}", errorMessage)
                } else {
                    errors.add(errorMessage)
                }
            }
        }

        for (int i = 0; i < stages.size(); i++) {
            Map config = loadEffectiveStageConfig(script, stages[i])
            if (!config.containsKey(requiredConfigKey)) {
                String errorMessage = "Your pipeline configuration does not contain the configuration " +
                    "key ${requiredConfigKey} for the stage ${stages[i]}. " + customMessage
                if (warnInsteadOfError) {
                    addPipelineWarning(script, "Missing configuration key ${requiredConfigKey}", errorMessage)
                } else {
                    errors.add(errorMessage)
                }
            }
        }

        if (general) {
            Map config = loadEffectiveGeneralConfig(script)
            if (!config.containsKey(requiredConfigKey)) {
                String errorMessage = "Your pipeline configuration does not contain the configuration " +
                    "key ${requiredConfigKey} in the general section. " + customMessage
                if (warnInsteadOfError) {
                    addPipelineWarning(script, "Missing configuration key ${requiredConfigKey}", errorMessage)
                } else {
                    errors.add(errorMessage)
                }
            }
        }

        if (postAction) {
            Map config = loadEffectivePostActionConfig(script)
            if (!config.containsKey(requiredConfigKey)) {
                String errorMessage = "Your pipeline configuration does not contain the configuration " +
                    "key ${requiredConfigKey} in the postActions section. " + customMessage
                if (warnInsteadOfError) {
                    addPipelineWarning(script, "Missing configuration key ${requiredConfigKey}", errorMessage)
                } else {
                    errors.add(errorMessage)
                }
            }
        }
    }
    return errors
}

static List checkForRemovedOrReplacedSteps(Script script, Map configChanges) {
    List errors = []
    configChanges.each { oldConfigKey, changes ->
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
            errors.add("Your pipeline configuration contains configuration for the step ${oldConfigKey}. " +
                "This step has been removed. " + customMessage)
        }
    }
    return errors
}

static List checkForRemovedOrReplacedStages(Script script, Map configChanges) {
    List errors = []
    configChanges.each { oldConfigKey, changes ->
        String customMessage = changes?.customMessage ?: ""
        String newStageName = changes?.newStageName ?: ""

        if (newStageName) {
            customMessage = "Please configure the stage ${newStageName} instead. " + customMessage
        }

        if (loadEffectiveStageConfig(script, oldConfigKey)) {
            errors.add("Your pipeline configuration contains configuration for the stage ${oldConfigKey}. " +
                "This stage has been removed. " + customMessage)
        }
    }
    return errors
}

static List checkForParameterTypeChanged(Script script, Map configChanges) {
    List errors = []
    configChanges.each { oldConfigKey, changes ->
        String oldType = changes?.oldType ?: ""
        String newType = changes?.newType ?: ""
        List steps = changes?.steps ?: []
        List stages = changes?.stages ?: []
        Boolean general = changes?.general ?: false
        String customMessage = changes?.customMessage ?: ""

        if (oldType != "String") {
            errors.add("Your legacy config settings contain an entry for parameterTypeChanged with the key ${oldConfigKey} with the unsupported type ${oldType}. " +
                "Currently only the type 'String' is supported.")
            return
        }

        for (int i = 0; i < steps.size(); i++) {
            Map config = loadEffectiveStepConfig(script, steps[i])
            if (config.containsKey(oldConfigKey)) {
                if (oldType == "String" && config.get(oldConfigKey) instanceof String) {
                    errors.add("Your pipeline configuration contains the configuration key ${oldConfigKey} for the step ${steps[i]}. " +
                        "The type of this configuration parameter was changed from ${oldType} to ${newType}. " + customMessage)
                }
            }
        }

        for (int i = 0; i < stages.size(); i++) {
            Map config = loadEffectiveStageConfig(script, stages[i])
            if (config.containsKey(oldConfigKey)) {
                if (oldType == "String" && config.get(oldConfigKey) instanceof String) {
                    errors.add("Your pipeline configuration contains the configuration key ${oldConfigKey} for the stage ${stages[i]}. " +
                        "The type of this configuration parameter was changed from ${oldType} to ${newType}. " + customMessage)
                }
            }
        }

        if (general) {
            Map config = loadEffectiveGeneralConfig(script)
            if (config.containsKey(oldConfigKey)) {
                if (oldType == "String" && config.get(oldConfigKey) instanceof String) {
                    errors.add("Your pipeline configuration contains the configuration key ${oldConfigKey} in the general section. " +
                        "The type of this configuration parameter was changed from ${oldType} to ${newType}. " + customMessage)
                }
            }
        }
    }
    return errors
}

static List checkForRenamedNpmScripts(Script script, Map configChanges) {
    List errors = []
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
                errors.add(errorMessage)
            }
        }
    }
    return errors
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
