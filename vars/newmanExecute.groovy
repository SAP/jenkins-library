import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = 'newmanExecute'
@Field Set STEP_CONFIG_KEYS = [
    'dockerImage',
    'failOnError',
    'gitBranch',
    'gitSshKeyCredentialsId',
    'newmanCollection',
    'newmanEnvironment',
    'newmanGlobals',
    'newmanRunCommand',
    'stashContent',
    'testRepository'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]
        def utils = parameters?.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME], config)

        config.stashContent = config.testRepository
            ?[GitUtils.handleTestRepository(this, config)]
            :utils.unstashAll(config.stashContent)

        List collectionList = findFiles(glob: config.newmanCollection)?.toList()
        if (collectionList.isEmpty()) {
            error "[${STEP_NAME}] No collection found with pattern '${config.newmanCollection}'"
        } else {
            echo "[${STEP_NAME}] Found files ${collectionList}"
        }

        dockerExecute(
            dockerImage: config.dockerImage,
            stashContent: config.stashContent
        ) {
            sh 'npm install newman --global --quiet'
            for(String collection : collectionList){
                def collectionDisplayName = collection.toString().replace(File.separatorChar,(char)'_').tokenize('.').first()
                // resolve templates
                def command = SimpleTemplateEngine.newInstance()
                    .createTemplate(config.newmanRunCommand)
                    .make([
                        config: config.plus([newmanCollection: collection]),
                        collectionDisplayName: collectionDisplayName
                    ]).toString()
                if(!config.failOnError) command += ' --suppress-exit-code'
                sh "newman ${command}"
            }
        }
    }
}
