import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** @see dockerExecute */
    'stashContent',
    /** Defines the environment variables to pass to the test execution.*/
    'envVars',
    /** Defines the behavior, in case tests fail. For example, in case of `outputFormat: 'junit'` you should set it to `false`. Otherwise test results cannot be recorded using the `testsPublishhResults` step afterwards.*/
    'failOnError',
    /**
     * Defines the format of the test result output. `junit` would be the standard for automated build environments but you could use also the option `tap`.
     * @possibleValues `junit`, `tap`
     */
    'outputFormat',
    /**
     * Defines the version of **bats-core** to be used. By default we use the version from the master branch.
     */
    'repository',
    /** For the transformation of the test result to xUnit format the node module **tap-xunit** is used. This parameter defines the name of the test package used in the xUnit result file.*/
    'testPackage',
    /** Defines either the directory which contains the test files (`*.bats`) or a single file. You can find further details in the [Bats-core documentation](https://github.com/bats-core/bats-core#usage).*/
    'testPath',
    /** Allows to load tests from another repository.*/
    'testRepository',
    /** Defines the branch where the tests are located, in case the tests are not located in the master branch.*/
    'gitBranch',
    /**
     * Defines the access credentials for protected repositories.
     * Note: In case of using a protected repository, `testRepository` should include the ssh link to the repository.
     */
    'gitSshKeyCredentialsId'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/** This step executes tests using the [Bash Automated Testing System - bats-core](https://github.com/bats-core/bats-core)*/
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils ?: new Utils()

        def script = checkScript(this, parameters) ?: this

        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // report to SWA
        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        InfluxData.addField('step_data', 'bats', false)

        config.stashContent = config.testRepository
            ?[GitUtils.handleTestRepository(this, config)]
            :utils.unstashAll(config.stashContent)

        //resolve commonPipelineEnvironment references in envVars
        config.envVarList = []
        config.envVars.each {e ->
            def envValue = GStringTemplateEngine.newInstance().createTemplate(e.getValue()).make(commonPipelineEnvironment: script.commonPipelineEnvironment).toString()
            config.envVarList.add("${e.getKey()}=${envValue}")
        }

        withEnv(config.envVarList) {
            sh "git clone ${config.repository}"
            try {
                sh "bats-core/bin/bats --recursive --tap ${config.testPath} > 'TEST-${config.testPackage}.tap'"
                InfluxData.addField('step_data', 'bats', true)
            } catch (err) {
                echo "[${STEP_NAME}] One or more tests failed"
                if (config.failOnError) throw err
            } finally {
                sh "cat 'TEST-${config.testPackage}.tap'"
                if (config.outputFormat == 'junit') {
                    dockerExecute(
                        script: script,
                        dockerImage: config.dockerImage,
                        dockerEnvVars: config.dockerEnvVars,
                        dockerOptions: config.dockerOptions,
                        dockerWorkspace: config.dockerWorkspace,
                        stashContent: config.stashContent
                    ) {
                        sh "NPM_CONFIG_PREFIX=~/.npm-global npm install tap-xunit -g"
                        sh "cat 'TEST-${config.testPackage}.tap' | PATH=\$PATH:~/.npm-global/bin tap-xunit --package='${config.testPackage}' > TEST-${config.testPackage}.xml"
                    }
                }
            }
        }
    }
}
