import hudson.AbortException

import com.sap.piper.DockerUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the tool used for the build.
     * @possibleValues `docker`, `kaniko`, `maven`, `mta`, `npm`
     */
    'buildTool',
    /** For Docker builds only (mandatory): name of the image to be built. */
    'dockerImageName',
    /** For Docker builds only: Defines the registry url where the image should be pushed to, incl. the protocol like `https://my.registry.com`. If it is not defined, image will not be pushed to a registry.*/
    'dockerRegistryUrl',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([

    /** Only for Docker builds on the local daemon: Defines the build options for the build.*/
    'containerBuildOptions',
    /** For custom build types: Defines the command to be executed within the `dockerImage` in order to execute the build. */
    'dockerCommand',
    /** For custom build types: Image to be used for builds in case they should run inside a custom Docker container */
    'dockerImage',
    /** For Docker builds only (mandatory): tag of the image to be built. */
    'dockerImageTag',
    /** For buildTool npm: Execute npm install (boolean, default 'true') */
    'npmInstall',
    /** For buildTool npm: List of npm run scripts to execute */
    'npmRunScripts',
    /** Defines if a container image(s) should be created with Cloud Native Buildpacks using the artifact produced by the `buildTool`. */
    'cnbBuild',
    /** toggles if a helmExecute is triggered at end of the step after invoking the build tool  */
    'helmExecute'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step serves as generic entry point in pipelines for building artifacts.
 *
 * You can use pre-defined `buildTool`s.
 *
 * Alternatively you can define a command via `dockerCommand` which should be executed in `dockerImage`.<br />
 * This allows you to trigger any build tool using a defined Docker container which provides the required build infrastructure.
 *
 * When using `buildTool: docker` or `buildTool: kaniko` the created container image is uploaded to a container registry.<br />
 * You need to make sure that the required credentials are provided to the step.
 *
 * For all other `buildTool`s the artifact will just be stored in the workspace and could then be `stash`ed for later use.
 *
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME
        // handle deprecated parameters
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('dockerImageTag', script.commonPipelineEnvironment.getArtifactVersion())
            .addIfEmpty('buildTool', script.commonPipelineEnvironment.getBuildTool())
            .use()

        switch(config.buildTool){
            case 'maven':
                mavenBuild script: script
                // in case node_modules exists we assume npm install was executed by maven clean install
                if (fileExists('package.json') && !fileExists('node_modules')) {
                    npmExecuteScripts script: script, install: true
                }
                break
            case 'mta':
                mtaBuild script: script
                break
            case 'npm':
                npmExecuteScripts script: script, install: config.npmInstall, runScripts: config.npmRunScripts
                break
            case ['docker', 'kaniko']: //handled below
                break
            default:
                if (config.dockerImage && config.dockerCommand) {
                    dockerExecute(
                        script: script,
                        dockerImage: config.dockerImage,
                    ) {
                        sh "${config.dockerCommand}"
                    }
                } else {
                    error "[${STEP_NAME}] buildTool not set and no dockerImage & dockerCommand provided."
                }
        }
        if (config.cnbBuild) {
            if (config.buildTool in ['npm', 'gradle', 'maven', 'mta', 'docker']) {
                cnbBuild script: script
            } else {
                throw new AbortException("ERROR - 'cnbBuild' does not support '${config.buildTool}' as a buildTool.")
            }
        } else if (config.buildTool == 'kaniko' || config.buildTool == 'docker') {
            kanikoExecute script: script
        }
        if(config.helmExecute) {
          helmExecute script: script
        }
    }
}
