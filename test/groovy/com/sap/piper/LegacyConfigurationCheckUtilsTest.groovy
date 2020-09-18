package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.Rules

class LegacyConfigurationCheckUtilsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)


    @Before
    void init() {
        DefaultValueCache.createInstance([
            steps: [
                mavenExecute: [
                    dockerImage: 'maven:3.5-jdk-8-alpine'
                ]
            ]
        ])
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
        Map configChanges = [oldConfigKey: [steps: ['someStep'], newConfigKey: "newConfigKey",customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)

    }

    @Test
    void testCheckForRemovedConfigKeysWithWarning() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedStageConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedGeneralConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedPostActionConfigKeys() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key oldConfigKey for the step someStep. " +
            "This configuration option was removed. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [someStep: [oldConfigKey: false]]]
        Map configChanges = [oldConfigKey: [steps: ['someStep'], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedConfigKeys(nullScript, configChanges)
    }

    @Test
    void testCheckForRemovedStep() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the step oldStep. " +
            "This step has been removed. Please configure the step newStep instead. test")
        nullScript.commonPipelineEnvironment.configuration = [steps: [oldStep: [configKey: false]]]
        Map configChanges = [oldStep: [newStepName: 'newStep', customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedSteps(nullScript, configChanges)

    }

    @Test
    void testCheckForReplacedStep() {

    }

    @Test
    void testCheckForRemovedStepOnlyProjectConfig() {

    }

    @Test
    void testCheckForRemovedStage() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains configuration for the stage oldStage. " +
            "This stage has been removed. Please configure the stage newStage instead. test")
        nullScript.commonPipelineEnvironment.configuration = [stages: [oldStage: [configKey: false]]]
        Map configChanges = [oldStage: [newStageName: 'newStage', customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRemovedOrReplacedStages(nullScript, configChanges)

    }

    @Test
    void testCheckForReplacedStage() {

    }

    @Test
    void testCheckForParameterTypeChanged() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your pipeline configuration contains the configuration key configKeyOldType for the stage productionDeployment. " +
            "The type of this configuration parameter was changed from String to List. test")
        nullScript.commonPipelineEnvironment.configuration = [stages: [productionDeployment: [configKeyOldType: "string"]]]
        Map configChanges = [configKeyOldType: [oldType: "String", newType: "List", stages: ["productionDeployment", "endToEndTests"], customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForParameterTypeChanged(nullScript, configChanges)

    }

    @Test
    void testCheckForStageParameterTypeChanged() {

    }

    @Test
    void testCheckForGeneralParameterTypeChanged() {

    }

    @Test
    void testCheckForParameterUnsupportedTypeChanged() {

    }

    @Test
    void testCheckForRenamedNpmScripts() {
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

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("Your package.json file package.json contains an npm script using the deprecated name oldNpmScriptName. " +
            "Please rename the script to newNpmScriptName, since the script oldNpmScriptName will not be executed by the pipeline anymore. test")
        nullScript.commonPipelineEnvironment.configuration = [stages: [productionDeployment: [configKeyOldType: "string"]]]
        Map configChanges = [oldNpmScriptName: [newScriptName: "newNpmScriptName", customMessage: "test"]]

        LegacyConfigurationCheckUtils.checkForRenamedNpmScripts(nullScript, configChanges)
    }

    @Test
    void testCheckForRenamedNpmScriptsWithWarning() {

    }
}
