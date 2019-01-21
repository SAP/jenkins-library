import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    /**
      * Docker image for code execution.
      */
    'dockerImage',
    /**
      * Defines the behavior, in case tests fail.
      * @possibleValues `true`, `false`
      */
    'failOnError',
    /**
      * see `testRepository`
      */
    'gitBranch',
    /**
      * see `testRepository`
      */
    'gitSshKeyCredentialsId',
    /**
      * The test collection that should be executed. This could also be a file pattern.
      */
    'newmanCollection',
    /**
      * Specify an environment file path or URL. Environments provide a set of variables that one can use within collections.
      * see also [Newman docs](https://github.com/postmanlabs/newman#newman-run-collection-file-source-options)
      */
    'newmanEnvironment',
    /**
      * Specify the file path or URL for global variables. Global variables are similar to environment variables but have a lower precedence and can be overridden by environment variables having the same name.
      * see also [Newman docs](https://github.com/postmanlabs/newman#newman-run-collection-file-source-options)
      */
    'newmanGlobals',
    /**
      * The shell command that will be executed inside the docker container to install Newman.
      */
    'newmanInstallCommand',
    /**
      * The newman command that will be executed inside the docker container.
      */
    'newmanRunCommand',
    /**
      * If specific stashes should be considered for the tests, you can pass this via this parameter.
      */
    'stashContent',
    /**
      * In case the test implementation is stored in a different repository than the code itself, you can define the repository containing the tests using parameter `testRepository` and if required `gitBranch` (for a different branch than master) and `gitSshKeyCredentialsId` (for protected repositories).
      * For protected repositories the `testRepository` needs to contain the ssh git url.
      */
    'testRepository'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = parameters?.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'script missing',
            stepParam1: parameters?.script == null
        ], config)

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
            script: script,
            dockerImage: config.dockerImage,
            stashContent: config.stashContent
        ) {
            sh "NPM_CONFIG_PREFIX=~/.npm-global ${config.newmanInstallCommand}"
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
                sh "PATH=\$PATH:~/.npm-global/bin newman ${command}"
            }
        }
    }
}
