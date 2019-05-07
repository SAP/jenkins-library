import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the tool which is used for building the artifact.<br />
     * Currently, it is possible to select two behaviors of the step:
     *
     * 1. Golang-specific behavior (`buildTool: golang`). Assumption here is that project uses the dependency management tool _dep_
     * 2. Custom-specific behavior for all other values of `buildTool`
     *
     * @possibleValues `golang`, any other build tool
     */
    'buildTool',
    'detect',
    /**
     * Name of the Synopsis Detect (formerly BlackDuck) project.
     * @parentConfigKey detect
     */
    'projectName',
    /**
     * Version of the Synopsis Detect (formerly BlackDuck) project.
     * @parentConfigKey detect
     */
    'projectVersion',
    /**
     * Properties passed to the Synopsis Detect (formerly BlackDuck) scan. You can find details in the [Synopsis Detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/622846/Using+Synopsys+Detect+Properties)
     * @parentConfigKey detect
     */
    'scanProperties',
    /**
     * Server url to the Synopsis Detect (formerly BlackDuck) Server.
     * @parentConfigKey detect
     */
    'serverUrl',
    /**
     * Jenkins 'Secret text' credentials ID containing the API token used to authenticate with the Synopsis Detect (formerly BlackDuck) Server.
     * @parentConfigKey detect
     */
    'userTokenCredentialsId'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** If specific stashes should be considered for the scan, their names need to be passed via the parameter `stashContent`. */
    'stashContent'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@Field Map CONFIG_KEY_COMPATIBILITY = [
    detect: [
        projectName: 'projectName',
        projectVersion: 'projectVersion',
        scanProperties: 'scanProperties',
        serverUrl: 'serverUrl',
        userTokenCredentialsId: 'userTokenCredentialsId'
    ]
]

/**
 * This step executes [Synopsis Detect](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/62423113/Synopsys+Detect) scans.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS,CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .dependingOn('buildTool').mixin('dockerImage')
            .dependingOn('buildTool').mixin('dockerWorkspace')
            .withMandatoryProperty('detect/userTokenCredentialsId')
            .withMandatoryProperty('detect/projectName')
            .withMandatoryProperty('detect/projectVersion')
            .use()

        config.stashContent = utils.unstashAll(config.stashContent)

        script.commonPipelineEnvironment.setInfluxStepData('detect', false)

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'buildTool',
            stepParam1: config.buildTool ?: 'default'
        ], config)

        //prepare Hub Detect execution using package manager
        switch (config.buildTool) {
            case 'golang':
                dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                    sh 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
                    sh 'ln --symbolic $(pwd) $GOPATH/src/hub'
                    sh 'cd $GOPATH/src/hub && dep ensure'
                }
                break
            default:
                //no additional tasks are performed
                echo "[${STEP_NAME}] No preparation steps performed for scan. Please make sure to properly set configuration for `detect.scanProperties`"
        }

        withCredentials ([string(
            credentialsId: config.detect.userTokenCredentialsId,
            variable: 'detectApiToken'
        )]) {
            config.detect.scanProperties += [
                "--blackduck.api.token=${detectApiToken}",
                "--detect.project.name='${config.detect.projectName}'",
                "--detect.project.version.name='${config.detect.projectVersion}'",
                "--detect.code.location.name='${config.detect.projectName}/${config.detect.projectVersion}'",
                "--blackduck.url=${config.detect.serverUrl}",

                //ToDo: get format of paths -> add another config???

                "--detect.blackduck.signature.scanner.paths=.",
            ]

            def detectProperties = config.detect.scanProperties.join(' ')
            echo "[${STEP_NAME}] Running with following Detect configuration: ${detectProperties}"
            synopsys_detect detectProperties
            script.commonPipelineEnvironment.setInfluxStepData('detect', true)
        }
    }
}
