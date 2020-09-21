package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.Rules

import static org.junit.Assert.assertEquals


class LegacyConfigurationCheckUtilsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    String echoOutput = ""

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)


    @Before
    void init() {
        DefaultValueCache.createInstance([
            steps: [
                customStep: [
                    param: 'test'
                ]
            ]
        ])
        helper.registerAllowedMethod('addBadge', [Map], {return})
        helper.registerAllowedMethod('createSummary', [Map], {return})
        helper.registerAllowedMethod("echo", [String.class], { s -> echoOutput = s})
        helper.registerAllowedMethod('findFiles', [Map], {m ->
            if(m.glob == '**/package.json') {
                return [new File("package.json")].toArray()
            } else {
                return []
            }
        })
        helper.registerAllowedMethod('readJSON', [Map], { m ->
            if (m.file.contains('package.json')) {
                return [scripts: [oldNpmScriptName: "echo test",
                                  npmScript2: "echo test"]]
            } else {
                return [:]
            }
        })
    }

    @Test
    void testCheckForRemovedConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
        "This configuration option was removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForReplacedConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. Please use the parameter newConfigKey instead. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], newConfigKey: "newConfigKey", customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)

    }

    @Test
    void testCheckForRemovedConfigKeysWithWarning() {
        String expectedWarning = "[WARNING] Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. test"

        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], warnInsteadOfError: true, customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
        assertEquals(expectedWarning, echoOutput)
    }

    @Test
    void testCheckForRemovedStageConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the stage someStage. " +
            "This configuration option was removed. ")
        nullScript.commonPipelineEnvironment.configuration = [stages: [someStage: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [stages: ['someStage']]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedGeneralConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey in the general section. " +
            "This configuration option was removed. ")
        nullScript.commonPipelineEnvironment.configuration = [general: [oldConfigKey: false]]
        Map configChanges = [oldConfigKey: [general: true]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedPostActionConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey in the postActions section. " +
            "This configuration option was removed. ")
        nullScript.commonPipelineEnvironment.configuration = [postActions: [oldConfigKey: false]]
        Map configChanges = [oldConfigKey: [postAction: true]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForReplacedStep() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the step oldStep. " +
            "This step has been removed. Please configure the step newStep instead. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [oldStep: [configKey: false]]]
        Map configChanges = [oldStep: [newStepName: 'newStep', customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedSteps(nullScript, configChanges)

    }

    @Test
    void testCheckForRemovedStep() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the step oldStep. " +
            "This step has been removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [oldStep: [configKey: false]]]
        Map configChanges = [oldStep: [customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedSteps(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedStepOnlyProjectConfig() {
        DefaultValueCache.createInstance([
            steps: [
                oldStep: [
                    configKey: false
                ]
            ]
        ])
        nullScript.commonPipelineEnvironment.configuration = [steps: [newStep: [configKey: false]]]
        Map configChanges = [oldStep: [onlyCheckProjectConfig: true]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedSteps(nullScript, configChanges)
    }

    @Test
    void testCheckForReplacedStage() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the stage oldStage. " +
            "This stage has been removed. Please configure the stage newStage instead. test")
        nullScript.commonPipelineEnvironment.configuration = [stages: [oldStage: [configKey: false]]]
        Map configChanges = [oldStage: [newStageName: 'newStage', customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedStages(nullScript, configChanges)

    }

    @Test
    void testCheckForRemovedStage() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the stage oldStage. " +
            "This stage has been removed. ")
        nullScript.commonPipelineEnvironment.configuration = [stages: [oldStage: [configKey: false]]]
        Map configChanges = [oldStage: []]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedStages(nullScript, configChanges)
    }

    @Test
    void testCheckForStageParameterTypeChanged() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key configKeyOldType for the stage productionDeployment. " +
            "The type of this configuration parameter was changed from String to List. test")
        nullScript.commonPipelineEnvironment.configuration = [stages: [productionDeployment: [configKeyOldType: "string"]]]
        Map configChanges = [configKeyOldType: [oldType: "String", newType: "List", stages: ["productionDeployment", "endToEndTests"], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForParameterTypeChanged(nullScript, configChanges)

    }

    @Test
    void testCheckForStepParameterTypeChanged() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key configKeyOldType for the step testStep. " +
            "The type of this configuration parameter was changed from String to List. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [testStep: [configKeyOldType: "string"]]]
        Map configChanges = [configKeyOldType: [oldType: "String", newType: "List", steps: ["testStep"], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForParameterTypeChanged(nullScript, configChanges)
    }

    @Test
    void testCheckForGeneralParameterTypeChanged() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key configKeyOldType in the general section. " +
            "The type of this configuration parameter was changed from String to List. test")
        nullScript.commonPipelineEnvironment.configuration = [general: [configKeyOldType: "string"]]
        Map configChanges = [configKeyOldType: [oldType: "String", newType: "List", general: true, customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForParameterTypeChanged(nullScript, configChanges)
    }

    @Test
    void testCheckForUnsupportedParameterTypeChanged() {
        String expectedWarning = "Your legacy config settings contain an entry for parameterTypeChanged with the key configKeyOldType with the unsupported type Map. " +
            "Currently only the type 'String' is supported."
        nullScript.commonPipelineEnvironment.configuration = [steps: [testStep: [configKeyOldType: [test: true]]]]
        Map configChanges = [configKeyOldType: [oldType: "Map", newType: "List", steps: ["testStep"], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForParameterTypeChanged(nullScript, configChanges)
        assertEquals(expectedWarning, echoOutput)
    }

    @Test
    void testCheckForRenamedNpmScripts() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your package.json file package.json contains an npm script using the deprecated name oldNpmScriptName. " +
            "Please rename the script to newNpmScriptName, since the script oldNpmScriptName will not be executed by the pipeline anymore. test")
        Map configChanges = [oldNpmScriptName: [newScriptName: "newNpmScriptName", customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRenamedNpmScripts(nullScript, configChanges)
    }

    @Test
    void testCheckForRenamedNpmScriptsWithWarning() {
        Map configChanges = [oldNpmScriptName: [newScriptName: "newNpmScriptName", warnInsteadOfError: true, customMessage: "test"]]
        String expectedWarning = "[WARNING] Your package.json file package.json contains an npm script using the deprecated name oldNpmScriptName. " +
            "Please rename the script to newNpmScriptName, since the script oldNpmScriptName will not be executed by the pipeline anymore. test"

        LegacyConfigurationCheckUtils.checkForRenamedNpmScripts(nullScript, configChanges)

        assertEquals(expectedWarning, echoOutput)
    }

    @Test
    void testCheckConfigurationRemovedOrReplacedConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [
            removedOrReplacedConfigKeys: [
                oldConfigKey: [
                    steps: ['someStep'],
                    customMessage: "test"
                ]
            ]
        ]

        LegacyConfigurationCheckUtils.checkConfiguration(nullScript, configChanges)
    }

    @Test
    void testCheckConfigurationRemovedOrReplacedSteps() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the step oldStep. " +
            "This step has been removed. Please configure the step newStep instead. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [oldStep: [configKey: false]]]
        Map configChanges = [
            removedOrReplacedSteps: [
                oldStep: [
                    newStepName: 'newStep',
                    customMessage: "test"
                ]
            ]
        ]

        LegacyConfigurationCheckUtils.checkConfiguration(nullScript, configChanges)
    }

    @Test
    void testCheckConfigurationRemovedOrReplacedStages() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the stage oldStage. " +
            "This stage has been removed. ")
        nullScript.commonPipelineEnvironment.configuration = [stages: [oldStage: [configKey: false]]]
        Map configChanges = [
            removedOrReplacedStages: [
                oldStage: []
            ]
        ]

        LegacyConfigurationCheckUtils.checkConfiguration(nullScript, configChanges)
    }

    @Test
    void testCheckConfigurationParameterTypeChanged() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key configKeyOldType for the step testStep. " +
            "The type of this configuration parameter was changed from String to List. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [testStep: [configKeyOldType: "string"]]]
        Map configChanges = [
            parameterTypeChanged: [
                configKeyOldType: [
                    oldType: "String",
                    newType: "List",
                    steps: ["testStep"],
                    customMessage: "test"]
            ]
        ]

        LegacyConfigurationCheckUtils.checkConfiguration(nullScript, configChanges)
    }

    @Test
    void testCheckConfigurationRenamedNpmScript() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your package.json file package.json contains an npm script using the deprecated name oldNpmScriptName. " +
            "Please rename the script to newNpmScriptName, since the script oldNpmScriptName will not be executed by the pipeline anymore. test")
        Map configChanges = [
            renamedNpmScript: [
                oldNpmScriptName: [
                    newScriptName: "newNpmScriptName",
                    customMessage: "test"]
            ]
        ]

        LegacyConfigurationCheckUtils.checkConfiguration(nullScript, configChanges)
    }
}
