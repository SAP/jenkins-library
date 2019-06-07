package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.DefaultValueCache

import util.BasePiperTest
import util.JenkinsFileExistsRule
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
        .around(new JenkinsFileExistsRule(this))

    @Before
    void init()  {

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(), [:])

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

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(),
        [
            stages: [
                testStage2: [testStage: 'myVal2'],
                testStage3: [testStage: 'myVal3']
            ]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                containsInAnyOrder(
                    'testStage2',
                    'testStage3'
                ),
                hasSize(2)
            )
        )
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

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(), [
            general: [testGeneral: 'myVal1'],
            stages: [testStage2: [testStage: 'myVal2']],
            steps: [thirdStep: [testStep: 'myVal3']]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                containsInAnyOrder(
                    'testStage1',
                    'testStage2',
                    'testStage3'
                ),
                hasSize(3)
            )
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.firstStep, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage2.secondStep, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage3.thirdStep, is(true))

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

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(), [
            general: [testGeneral: 'myVal1'],
            stages: [:],
            steps: [thirdStep: [testStep: 'myVal3']]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                containsInAnyOrder(
                    'testStage1',
                    'testStage3'
                ),
                hasSize(2)
            )
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.firstStep, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage2?.secondStep, is(false))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage3.thirdStep, is(true))

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

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(), [
            general: [myKey1_1: 'myVal1_1'],
            stages: [:],
            steps: [thirdStep: [myKey3_1: 'myVal3_1']]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                containsInAnyOrder(
                    'testStage1',
                    'testStage3'
                ),
                hasSize(2)
            )
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.firstStep, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage2?.secondStep, is(false))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage3.thirdStep, is(true))

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

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                contains('testStage1'),
                hasSize(1)
            )
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.firstStep, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.secondStep, is(false))

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

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(),
        [
            general: [:],
            stages: [testStage1: [myVal1: '**/conf.js']]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'testDefault.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                contains('testStage1'),
                hasSize(1)
            )
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.firstStep, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.testStage1.secondStep, is(false))
    }

    @Test
    void testVerboseOption() {
        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(),
        [
            general: [verbose: true],
            steps: [:],
            stages: [
                Test: [:],
                Integration: [test: 'test'],
                Acceptance: [test: 'test']
            ]
        ])

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

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(),
        [
            general: [:],
            steps: [:],
            stages: [
                Test: [:],
                Integration: [test: 'test'],
                Acceptance: [test: 'test']
            ]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.keySet(),
            allOf(
                containsInAnyOrder(
                    'Acceptance',
                    'Integration'
                ),
                hasSize(2)
            )
        )

    }

    @Test
    void testPiperStepActivation() {

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(), [
            general: [:],
            steps: [
                cloudFoundryDeploy: [cfSpace: 'myTestSpace'],
                newmanExecute: [newmanCollection: 'myCollection.json']
            ],
            stages: [:]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )

        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.Acceptance.cloudFoundryDeploy, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.Acceptance.newmanExecute, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.Acceptance.newmanExecute, is(true))
    }

    @Test
    void testPiperStepActivationWithStage() {

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(), [
            general: [:],
            steps: [:],
            stages: [Acceptance: [cfSpace: 'test']]
        ])

        jsr.step.piperInitRunStageConfiguration(
            script: nullScript,
            juStabUtils: utils,
            stageConfigResource: 'com.sap.piper/pipeline/stageDefaults.yml'
        )
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStep.Acceptance.cloudFoundryDeploy, is(true))
        assertThat(DefaultValueCache.getInstance().getProjectConfig().runStage.Acceptance, is(true))

    }
}
