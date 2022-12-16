package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules
import util.JenkinsShellCallRule
import com.sap.piper.PiperGoUtils

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class PiperInitRunStageConfigurationTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this)
    private ExpectedException thrown = new ExpectedException()
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private PiperGoUtils piperGoUtils = new PiperGoUtils(utils) { void unstashPiperBin() { }}
    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jryr)
        .around(thrown)
        .around(shellCallRule)
        .around(jlr)
        .around(jsr)
    @Before
    void init()  {

        binding.variables.env.STAGE_NAME = 'Test'

        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            switch (map.glob) {
                case '**/conf.js':
                    return [new File('conf.js')].toArray()
                case 'myCollection.json':
                    return [new File('myCollection.json')].toArray()
                default:
                    return [].toArray()
            }
        })
        helper.registerAllowedMethod("writeFile", [Map.class], null)
    }

    @Test
    void testVerboseOption() {
        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 0)
         helper.registerAllowedMethod("readJSON", [Map.class], { m ->
                     if (m.containsValue(".pipeline/stage_out.json")) {
                         return ["testStage1":false]
                     } else {
                         if (m.containsValue(".pipeline/step_out.json")) {
                             return  [:]
                         }
                         return [:]
                     }
                 })

         nullScript.commonPipelineEnvironment.configuration = [
            general: [verbose: true],
            steps: [:],
            stages: [
                Test: [:],
                Integration: [test: 'test'],
                Acceptance: [test: 'test']
            ]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(jlr.log, allOf(
            containsString('[piperInitRunStageConfiguration] Debug - Run Stage Configuration:'),
            containsString('[piperInitRunStageConfiguration] Debug - Run Step Configuration:')
        ))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))
    }

    @Test(expected = Exception.class)
    void testPiperShFailed() {
        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 1)
        helper.registerAllowedMethod("readJSON", [Map.class], { m ->
            if (m.containsValue(".pipeline/stage_out.json")) {
                return ["Integration":true, "Acceptance":true]
            } else {
                if (m.containsValue(".pipeline/step_out.json")) {
                    return  [ Integration: [test: true], Acceptance: [test: true]]
                }
                return [:]
            }
        })

        helper.registerAllowedMethod("findFiles", [Map.class], { map -> [].toArray() })

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )
    }
    
    @Test
    void testPiperInitDefault() {

        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 0)

        helper.registerAllowedMethod("readJSON", [Map.class], { m ->
            if (m.containsValue(".pipeline/stage_out.json")) {
                return ["Integration":true, "Acceptance":true]
            } else {
                if (m.containsValue(".pipeline/step_out.json")) {
                    return  [ Integration: [test: true], Acceptance: [test: true]]
                }
                return [:]
            }
        })

        helper.registerAllowedMethod("findFiles", [Map.class], { map -> [].toArray() })

        nullScript.commonPipelineEnvironment.configuration = [
            general: [:],
            steps: [:],
            stages: [
                Test: [:],
                Integration: [test: 'test'],
                Acceptance: [test: 'test']
            ]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.Acceptance, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.Integration, is(true))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))

    }

    @Test
    void testConditionOnlyProductiveBranchOnNonProductiveBranch() {
        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 0)
        helper.registerAllowedMethod("readJSON", [Map.class], { m ->
                    if (m.containsValue(".pipeline/stage_out.json")) {
                        return ["testStage1":true]
                    } else {
                        if (m.containsValue(".pipeline/step_out.json")) {
                            return  [:]
                        }
                        return [:]
                    }
                })

        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
spec:
  stages:
    - name: testStage1
      displayName: testStage1
      steps:
        - name: firstStep
          conditions:
          - filePattern: \'**/conf.js\'
'''
            } else {
                return '''
general: {}
steps: {}
stages:
  testStage1:
    runInAllBranches: false
'''
            }
        })

        binding.variables.env.BRANCH_NAME = 'test'

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stageConfigResource: 'testDefault.yml',
            productiveBranch: 'master'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(false))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))
    }

    @Test
    void testConditionOnlyProductiveBranchOnProductiveBranch() {
        helper.registerAllowedMethod("readJSON", [Map.class], { m ->
                    if (m.containsValue(".pipeline/stage_out.json")) {
                        return ["testStage1":true]
                    } else {
                        if (m.containsValue(".pipeline/step_out.json")) {
                            return  ["testStage1":["firstStep":true]]
                        }
                        return [:]
                    }
                })
        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 0)

        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
spec:
  stages:
    - name: testStage1
      displayName: testStage1
      steps:
        - name: firstStep
          conditions:
            - filePattern: \'**/conf.js\'
'''
            } else {
                return '''
general: {}
steps: {}
stages:
  testStage1:
    runInAllBranches: false
'''
            }
        })

        binding.variables.env.BRANCH_NAME = 'test'

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stageConfigResource: 'testDefault.yml',
            productiveBranch: 'test'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))
    }

    @Test
    void testStageExtensionExists() {
        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 0)
        helper.registerAllowedMethod("readJSON", [Map.class], { m ->
                    if (m.containsValue(".pipeline/stage_out.json")) {
                        return ["testStage1":false, "testStage2":false, "testStage3":false, "testStage4":false, "testStage5":false]
                    } else {
                        if (m.containsValue(".pipeline/step_out.json")) {
                            return  [:]
                        }
                        return [:]
                    }
                })

        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
spec:
  stages:
    - name: testStage1
      displayName: testStage1
      extensionExists: true
    - name: testStage2
      displayName: testStage2
      extensionExists: true
    - name: testStage3
      displayName: testStage3
      extensionExists: false
    - name: testStage4
      displayName: testStage4
      extensionExists: 'false'
    - name: testStage5
      displayName: testStage5
      dummy: true
'''
            } else {
                return '''
general:
  projectExtensionsDirectory: './extensions/'
steps: {}
'''
            }
        })

        helper.registerAllowedMethod('fileExists', [String], {path ->
            switch (path) {
                case './extensions/testStage1.groovy':
                    return true
                case './extensions/testStage2.groovy':
                    return false
                case './extensions/testStage3.groovy':
                    return true
                case './extensions/testStage4.groovy':
                    return true
                case './extensions/testStage5.groovy':
                    return true
                default:
                    return false
            }
        })

        nullScript.piperStageWrapper = [:]
        nullScript.piperStageWrapper.allowExtensions = {script -> return true}

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage2, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage3, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage4, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage5, is(false))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))
    }
}
