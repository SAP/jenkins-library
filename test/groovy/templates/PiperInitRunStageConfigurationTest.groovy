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

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class PiperInitRunStageConfigurationTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jryr)
        .around(thrown)
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
    }

    @Test
    void testStageConfig() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1: {}
  testStage2: {}
  testStage3: {}
'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        nullScript.commonPipelineEnvironment.configuration = [
            stages: [
                testStage2: [testStage: 'myVal2'],
                testStage3: [testStage: 'myVal3']
            ]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage2, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage3, is(true))
    }


    @Test
    void testConditionConfig() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        config: testGeneral
  testStage2:
    stepConditions:
      secondStep:
        config: testStage
  testStage3:
    stepConditions:
      thirdStep:
        config: testStep

'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        nullScript.commonPipelineEnvironment.configuration = [
            general: [testGeneral: 'myVal1'],
            stages: [testStage2: [testStage: 'myVal2']],
            steps: [thirdStep: [testStep: 'myVal3']]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage2, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage3, is(true))

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage2.secondStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage3.thirdStep, is(true))

    }

    @Test
    void testConditionConfigValue() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        config:
          testGeneral:
            - myValx
            - myVal1
  testStage2:
    stepConditions:
      secondStep:
        config:
          testStage:
            - maValXyz
  testStage3:
    stepConditions:
      thirdStep:
        config:
          testStep:
            - myVal3

'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        nullScript.commonPipelineEnvironment.configuration = [
            general: [testGeneral: 'myVal1'],
            stages: [:],
            steps: [thirdStep: [testStep: 'myVal3']]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage2, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage3, is(true))

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage2?.secondStep, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage3.thirdStep, is(true))

    }

    @Test
    void testConditionConfigKeys() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        configKeys:
          - myKey1_1
          - myKey1_2
  testStage2:
    stepConditions:
      secondStep:
        configKeys:
          - myKey2_1
  testStage3:
    stepConditions:
      thirdStep:
        configKeys:
          - myKey3_1
'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        nullScript.commonPipelineEnvironment.configuration = [
            general: [myKey1_1: 'myVal1_1'],
            stages: [:],
            steps: [thirdStep: [myKey3_1: 'myVal3_1']]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage2, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage3, is(true))

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage2?.secondStep, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage3.thirdStep, is(true))

    }


    @Test
    void testConditionFilePattern() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern: \'**/conf.js\'
      secondStep:
        filePattern: \'**/conf.jsx\'

'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.keySet(),
            allOf(
                contains('testStage1'),
                hasSize(1)
            )
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.secondStep, is(false))

    }

    @Test
    void testConditionFilePatternWithList() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePattern:
         - \'**/conf.js\'
         - \'myCollection.json\'
      secondStep:
        filePattern: \'**/conf.jsx\'

'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.secondStep, is(false))

    }

    @Test
    void testConditionFilePatternFromConfig() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        filePatternFromConfig: myVal1
      secondStep:
        filePatternFromConfig: myVal2

'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        nullScript.commonPipelineEnvironment.configuration = [
            general: [:],
            stages: [testStage1: [myVal1: '**/conf.js']]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.keySet(),
            allOf(
                contains('testStage1'),
                hasSize(1)
            )
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.secondStep, is(false))
    }

    @Test
    void testVerboseOption() {
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
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(jlr.log, allOf(
            containsString('[piperInitRunStageConfiguration] Debug - Run Stage Configuration:'),
            containsString('[piperInitRunStageConfiguration] Debug - Run Step Configuration:')
        ))
    }

    @Test
    void testPiperInitDefault() {

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
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.Acceptance, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.Integration, is(true))

    }

    @Test
    void testPiperStepActivation() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [:],
            steps: [
                cloudFoundryDeploy: [cfSpace: 'myTestSpace'],
                newmanExecute: [newmanCollection: 'myCollection.json']
            ],
            stages: [:]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.Acceptance.cloudFoundryDeploy, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.Acceptance.newmanExecute, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.Acceptance.newmanExecute, is(true))
    }

    @Test
    void testPiperStepActivationWithStage() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [:],
            steps: [:],
            stages: [Acceptance: [cfSpace: 'test']]
        ]

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.Acceptance.cloudFoundryDeploy, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.Acceptance, is(true))

    }

    @Test
    void testConditionNpmScripts() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        npmScripts: \'npmScript\'
      secondStep:
        filePattern: \'**/conf.jsx\'

'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        helper.registerAllowedMethod('findFiles', [Map], {m ->
            if(m.glob == '**/package.json') {
                return [new File("package.json")].toArray()
            } else {
                return []
            }
        })

        helper.registerAllowedMethod('readJSON', [Map], { m ->
            if (m.file == 'package.json') {
                return [scripts: [npmScript: "echo test",
                                  npmScript2: "echo test"]]
            } else {
                return [:]
            }
        })

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.secondStep, is(false))

    }

    @Test
    void testConditionNpmScriptsWithList() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
                    if(s == 'testDefault.yml') {
                        return '''
stages:
  testStage1:
    stepConditions:
      firstStep:
        npmScripts:
         - \'npmScript\'
         - \'npmScript2\'
      secondStep:
        filePattern: \'**/conf.jsx\'

'''
                    } else {
                        return '''
general: {}
steps: {}
'''
            }
        })

        helper.registerAllowedMethod('findFiles', [Map], {m ->
            if(m.glob == '**/package.json') {
                return [new File("package.json")].toArray()
            } else {
                return []
            }
        })

        helper.registerAllowedMethod('readJSON', [Map], { m ->
            if (m.file == 'package.json') {
                return [scripts: [npmScript: "echo test",
                                  npmScript2: "echo test"]]
            } else {
                return [:]
            }
        })

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.firstStep, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep.testStage1.secondStep, is(false))

    }

    @Test
    void testConditionOnlyProductiveBranchOnNonProductiveBranch() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    onlyProductiveBranch: true
    stepConditions:
      firstStep:
        filePattern: \'**/conf.js\'
'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        binding.variables.env.BRANCH_NAME = 'test'

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml',
            productiveBranch: 'master'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(false))
    }

    @Test
    void testConditionOnlyProductiveBranchOnProductiveBranch() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    onlyProductiveBranch: true
    stepConditions:
      firstStep:
        filePattern: \'**/conf.js\'
'''
            } else {
                return '''
general: {}
steps: {}
'''
            }
        })

        binding.variables.env.BRANCH_NAME = 'test'

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml',
            productiveBranch: 'test'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
    }

    @Test
    void testStageExtensionExists() {
        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if(s == 'testDefault.yml') {
                return '''
stages:
  testStage1:
    extensionExists: true
  testStage2:
    extensionExists: true
  testStage3:
    extensionExists: false
  testStage4:
    extensionExists: 'false'
  testStage5:
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
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage1, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage2, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage3, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage4, is(false))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStage.testStage5, is(false))
    }
}
