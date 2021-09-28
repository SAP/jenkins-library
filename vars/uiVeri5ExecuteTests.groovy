import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/uiVeri5ExecuteTests.yaml'

/*
 * Parameters read from config for backwards compatibility of groovy wrapper step:
 *
 * testRepository, gitBranch, gitSshKeyCredentialsId used for test repository loading
 */
@Field Set CONFIG_KEYS = [
    "gitBranch",
    "gitSshKeyCredentialsId",
    "testRepository",
]

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME
    def utils = parameters.juStabUtils ?: new Utils()
    Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, CONFIG_KEYS)
            .mixin(parameters, CONFIG_KEYS)
            .use()

    if (parameters.testRepository || config.testRepository ) {
        parameters.stashContent = [GitUtils.handleTestRepository(this, [gitBranch: config.gitBranch, gitSshKeyCredentialsId: config.gitSshKeyCredentialsId, testRepository: config.testRepository])]
    }

    List credentials = [
        [type: 'usernamePassword', id: 'seleniumHubCredentialsId', env: ['PIPER_SELENIUM_HUB_USER', 'PIPER_SELENIUM_HUB_PASSWORD']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
