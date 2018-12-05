import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'dockerImage',
    'dockerWorkspace',
    'envVars',
    'failOnError',
    'gitBranch',
    'gitSshKeyCredentialsId',
    'outputFormat',
    'repository',
    'stashContent',
    'testPackage',
    'testPath',
    'testRepository'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils ?: new Utils()

        def script = checkScript(this, parameters) ?: this

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // report to SWA
        utils.pushToSWA([step: STEP_NAME,
                        stepParam1: parameters?.script == null], config)

        script.commonPipelineEnvironment.setInfluxStepData('bats', false)

        config.stashContent = config.testRepository
            ?[GitUtils.handleTestRepository(this, config)]
            :utils.unstashAll(config.stashContent)

        //resolve commonPipelineEnvironment references in envVars
        config.envVarList = []
        config.envVars.each {e ->
            def envValue = SimpleTemplateEngine.newInstance().createTemplate(e.getValue()).make(commonPipelineEnvironment: script.commonPipelineEnvironment).toString()
            config.envVarList.add("${e.getKey()}=${envValue}")
        }

        withEnv(config.envVarList) {
            sh "git clone ${config.repository}"
            try {
                sh "bats-core/bin/bats --recursive --tap ${config.testPath} > 'TEST-${config.testPackage}.tap'"
                script.commonPipelineEnvironment.setInfluxStepData('bats', true)
            } catch (err) {
                echo "[${STEP_NAME}] One or more tests failed"
                if (config.failOnError) throw err
            } finally {
                sh "cat 'TEST-${config.testPackage}.tap'"
                if (config.outputFormat == 'junit') {
                    dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                        sh "npm install tap-xunit -g"
                        sh "cat 'TEST-${config.testPackage}.tap' | tap-xunit --package='${config.testPackage}' > TEST-${config.testPackage}.xml"
                    }
                }
            }
        }
    }
}
